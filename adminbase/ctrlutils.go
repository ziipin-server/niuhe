package adminbase

import (
	"reflect"
	"unicode"
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
