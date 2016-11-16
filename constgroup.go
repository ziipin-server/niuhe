package niuhe

import (
	"reflect"
	"strconv"
	"strings"
)

//------------------------------------------------------------------------------

type iIntConstItem interface {
	Int() int
}

type IntConstItem struct {
	Name  string
	Value int
}

func (item IntConstItem) Int() int {
	return item.Value
}

type IntConstGroup struct {
	items map[int]string
	keys  []int
}

func (g *IntConstGroup) addField(value int, name string) {
	if _, ok := g.items[value]; !ok {
		g.keys = append(g.keys, value)
	}
	g.items[value] = name
}

func (g *IntConstGroup) GetName(value int) string {
	name, _ := g.items[value]
	return name
}

func (g *IntConstGroup) GetChoices() map[int]string {
	r := make(map[int]string, len(g.keys))
	for _, k := range g.keys {
		r[k], _ = g.items[k]
	}
	return r
}

func GetIntValueByFieldName(g interface{}, fieldName string) (value int) {
	if v := getValueByFieldName(g, fieldName); v != nil {
		value, _ = v.(int)
	}
	return
}

//------------------------------------------------------------------------------

type iStringConstItem interface {
	String() string
}

type StringConstItem struct {
	Name, Value string
}

func (item StringConstItem) String() string {
	return item.Value
}

type StringConstGroup struct {
	items map[string]string
	keys  []string
}

func (g *StringConstGroup) addField(value, name string) {
	if _, ok := g.items[value]; !ok {
		g.keys = append(g.keys, value)
	}
	g.items[value] = name
}

func (g *StringConstGroup) GetName(value string) string {
	name, _ := g.items[value]
	return name
}

func (g *StringConstGroup) GetChoices() map[string]string {
	r := make(map[string]string, len(g.keys))
	for _, k := range g.keys {
		r[k], _ = g.items[k]
	}
	return r
}

func GetStringValueByFieldName(g interface{}, fieldName string) (value string) {
	if v := getValueByFieldName(g, fieldName); v != nil {
		value, _ = v.(string)
	}
	return
}

//==============================================================================

func InitConstGroup(group interface{}) {
	var baseVal reflect.Value
	gVal := reflect.ValueOf(group).Elem()
	gType := gVal.Type()
	if baseVal = gVal.FieldByName("IntConstGroup"); baseVal.IsValid() {
		base := &IntConstGroup{make(map[int]string), make([]int, 0)}
		baseVal.Set(reflect.ValueOf(base))
	} else if baseVal = gVal.FieldByName("StringConstGroup"); baseVal.IsValid() {
		base := &StringConstGroup{make(map[string]string), make([]string, 0)}
		baseVal.Set(reflect.ValueOf(base))
	} else {
		panic("unknown const group type!")
	}
	for i := 0; i < gType.NumField(); i++ {
		var valueVal, nameVal reflect.Value
		fieldMeta := gType.Field(i)
		fieldVal := gVal.Field(i)
		if !fieldVal.CanInterface() {
			continue
		}
		v := fieldVal.Interface()
		if _, ok := v.(iIntConstItem); ok {
			valueStr, nameStr := getTags(fieldMeta.Tag)
			value64, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				panic(err)
			}
			valueVal = reflect.ValueOf(int(value64))
			nameVal = reflect.ValueOf(nameStr)
			baseVal.Interface().(*IntConstGroup).addField(int(value64), nameStr)
		} else if _, ok := v.(iStringConstItem); ok {
			valueStr, nameStr := getTags(fieldMeta.Tag)
			valueVal = reflect.ValueOf(valueStr)
			nameVal = reflect.ValueOf(nameStr)
			baseVal.Interface().(*StringConstGroup).addField(valueStr, nameStr)
		} else {
			continue
		}
		switch fieldMeta.Type {
		case reflect.TypeOf(IntConstItem{}), reflect.TypeOf(StringConstItem{}):
		default:
			fieldVal = fieldVal.FieldByName("IntConstItem")
		}
		fieldVal.FieldByName("Value").Set(valueVal)
		fieldVal.FieldByName("Name").Set(nameVal)
	}
}

func InitIntConstGroup(group interface{}) { InitConstGroup(group) }

func getTags(tag reflect.StructTag) (string, string) {
	tagStr := tag.Get("const")
	if len(tagStr) > 0 {
		tags := strings.SplitN(tag.Get("const"), ",", 2)
		if len(tags) == 1 {
			tags = append(tags, "")
		}
		return strings.TrimSpace(tags[0]), strings.TrimSpace(tags[1])
	}
	return tag.Get("value"), tag.Get("name")
}

func getValueByFieldName(g interface{}, fieldName string) (value interface{}) {
	val := reflect.Indirect(reflect.ValueOf(g))
	fieldVal := val.FieldByName(fieldName)
	if fieldVal.IsValid() {
		valueVal := fieldVal.FieldByName("Value")
		if valueVal.IsValid() {
			value = valueVal.Interface()
		}
	}
	return
}
