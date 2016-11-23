package niuhe

import (
	"encoding/json"
	"fmt"
	"testing"
)

func assertTrue(t *testing.T, expr bool, msg string, msgArgs ...interface{}) {
	if !expr {
		t.Error(fmt.Sprintf(msg, msgArgs...))
	}
}

func printJson(v interface{}) {
	jb, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(jb))
}

func TestOpIntMarshalJSON(t *testing.T) {
	var data struct {
		Number OpInt `json:"n"`
	}
	assertTrue(t, !data.Number.Exists(), "data.Number should be not exists")
	assertTrue(t, 2 == data.Number.ValueDefault(2), "data.Number should be 2")
	data.Number.Set(10)
	assertTrue(t, data.Number.Exists(), "data.Number should be exists")
	assertTrue(t, 10 == data.Number.MustValue(), "data.Number should be 10")
	printJson(data)
	data.Number.Clear()
	assertTrue(t, !data.Number.Exists(), "data.Number should be not exists")
	assertTrue(t, 2 == data.Number.ValueDefault(2), "data.Number should be 2")
	printJson(data)
}

func TestOpIntUnmarshalJSON(t *testing.T) {
	var data struct {
		Number OpInt `json:"n"`
	}
	json.Unmarshal([]byte("{}"), &data)
	assertTrue(t, !data.Number.Exists(), "data.Number should be not exists")
	json.Unmarshal([]byte("{\"n\":2}"), &data)
	assertTrue(t, data.Number.Exists(), "data.Number should be exists")
	assertTrue(t, 2 == data.Number.MustValue(), "data.Number should be 2")
}
