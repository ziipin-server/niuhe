package adminbase

import (
	"reflect"
	"regexp"
	"strings"
)

var (
	pascalRe *regexp.Regexp
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
	parts := pascalRe.FindAllString(pascalName, -1)
	return strings.ToLower(strings.Join(parts, "_"))
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
	return nil
}

func init() {
	pascalRe = regexp.MustCompile("[A-Z][a-z]*|[^A-Za-z]+")
}
