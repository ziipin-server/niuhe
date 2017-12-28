package adminbase

import (
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/robertkrimen/otto"
)

func MakeColRenderer(modelField, colField string) ColRenderer {
	return ColRenderer{
		ColName:    colField,
		RenderFunc: RenderField(modelField),
	}
}

func MakeSameRenderer(modelField string) ColRenderer {
	return MakeColRenderer(modelField, modelField)
}

func pascalToSnake(pascalName string) string {
	in := []rune(pascalName)
	out := make([]rune, 0, 2*len(in))
	hasPrefixUpper := false
	isLastLetter := false
	var prefixBegin int
	for i, ch := range in {
		if unicode.IsUpper(ch) {
			if !hasPrefixUpper {
				prefixBegin = i
				hasPrefixUpper = true
			}
			isLastLetter = true
		} else if unicode.IsLetter(ch) {
			if hasPrefixUpper {
				if prefixBegin > 0 {
					out = append(out, rune('_'))
				}
				prefixEnd := i - 1
				if prefixBegin < prefixEnd {
					for j := prefixBegin; j < prefixEnd; j++ {
						out = append(out, unicode.ToLower(in[j]))
					}
					out = append(out, rune('_'))
				}
				out = append(out, unicode.ToLower(in[prefixEnd]))
				hasPrefixUpper = false
			}
			out = append(out, ch)
			isLastLetter = true
		} else {
			if isLastLetter {
				out = append(out, rune('_'))
			}
			isLastLetter = false
			out = append(out, ch)
		}
	}
	if hasPrefixUpper {
		prefixEnd := len(in)
		if prefixBegin > 0 {
			out = append(out, rune('_'))
		}
		for j := prefixBegin; j < prefixEnd; j++ {
			out = append(out, unicode.ToLower(in[j]))
		}
	}
	return string(out)
}

func MakeSnakeRenderers(model interface{}, skipModelFields ...string) []ColRenderer {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() != reflect.Ptr {
		panic("model sample must be a pointer")
	}
	modelType = modelType.Elem()
	if modelType.Kind() != reflect.Struct {
		panic("model sample must be a pointer to a struct")
	}
	numField := modelType.NumField()
	renderers := make([]ColRenderer, 0, numField)
	for i := 0; i < numField; i++ {
		fieldInfo := modelType.Field(i)
		fname := fieldInfo.Name
		skip := false
		for _, skipName := range skipModelFields {
			if skipName == fname {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		renderers = append(renderers, MakeColRenderer(fname, pascalToSnake(fname)))
	}
	return renderers
}

func MakeRenderersByRules(ruleSrc string) []ColRenderer {
	lines := strings.Split(ruleSrc, "\n")
	renderers := make([]ColRenderer, 0, len(lines))
	for _, line := range lines {
		line = strings.Trim(line, " \t\r")
		if line == "" {
			continue
		}
		if m, err := regexp.MatchString("^\\w+$", line); err == nil && m {
			renderers = append(renderers, MakeSameRenderer(line))
		} else if m, err := regexp.MatchString(`^\w+\s*:\s*\w+$`, line); err == nil && m {
			re := regexp.MustCompile(`(\w+)\s*:\s*(\w+)`)
			caps := re.FindStringSubmatch(line)
			if len(caps) != 3 {
				panic("parse error")
			}
			renderers = append(renderers, MakeColRenderer(caps[2], caps[1]))
		}
	}
	return renderers
}

func MakeRenderersByJS(jsSrc string) []ColRenderer {
	renderers := make([]ColRenderer, 0)
	vm := otto.New()
	jsDef, err := vm.Run("(" + jsSrc + ")")
	if err != nil {
		panic(err)
	}
	if !jsDef.IsObject() {
		panic("jsDef must be an object")
	}
	defObj := jsDef.Object()
	cols := defObj.Keys()
	for _, col := range cols {
		colVal, err := defObj.Get(col)
		if err != nil {
			panic(err)
		}
		if colVal.IsNull() {
			renderers = append(renderers, MakeSameRenderer(col))
		} else if colVal.IsString() {
			modelField, _ := colVal.ToString()
			renderers = append(renderers, MakeColRenderer(modelField, col))
		} else if colVal.IsFunction() {
			r := ColRenderer{
				ColName: col,
				RenderFunc: func(model interface{}) interface{} {
					ret, err := colVal.Call(otto.UndefinedValue(), model)
					if err != nil {
						return nil
					}
					retVal, retErr := ret.Export()
					if retErr != nil {
						return nil
					}
					return retVal
				},
			}
			renderers = append(renderers, r)
		} else {
			panic("invalid field type " + col)
		}
	}
	return renderers
}
