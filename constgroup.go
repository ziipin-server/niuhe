package niuhe

import (
	"fmt"
	"reflect"
	"strconv"
)

type IntConstItem struct {
	Name  string
	Value int
}

type IntConstGroup struct {
	items map[int]string
}

func (g *IntConstGroup) addField(value int, name string) {
	g.items[value] = name
}

func InitIntConstGroup(group interface{}) {
	gElemValue := reflect.ValueOf(group).Elem()
	gElemType := gElemValue.Type()
	baseValue := gElemValue.FieldByName("IntConstGroup")
	base := &IntConstGroup{make(map[int]string)}
	baseValue.Set(reflect.ValueOf(base))
	for i := 0; i < gElemType.NumField(); i++ {
		fieldMeta := gElemType.Field(i)
		if fieldMeta.Type != reflect.TypeOf(IntConstItem{}) {
			continue
		}
		name := fieldMeta.Tag.Get("name")
		value, err := strconv.ParseInt(fieldMeta.Tag.Get("value"), 10, 64)
		if err != nil {
			panic(err)
		}
		gElemValue.Field(i).FieldByName("Value").SetInt(value)
		gElemValue.Field(i).FieldByName("Name").SetString(name)
		base.addField((int)(value), name)
	}
}

func (g *IntConstGroup) GetName(value int) string {
	name, exists := g.items[value]
	if exists {
		return name
	} else {
		return ""
	}
}

func (g *IntConstGroup) MustGetName(value int) string {
	name, exists := g.items[value]
	if !exists {
		panic(fmt.Sprintf("MustGetName fail: Cannot find value %d", value))
	}
	return name
}
