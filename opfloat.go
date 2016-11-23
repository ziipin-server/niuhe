package niuhe

import (
	"fmt"
	"strconv"
)

type OpFloat struct {
	value *float64
	store float64
}

func (f *OpFloat) Exists() bool {
	return f.value != nil
}

func (f *OpFloat) Value() (float64, bool) {
	if !f.Exists() {
		return 0, false
	}
	return *f.value, true
}

func (f *OpFloat) MustValue() float64 {
	if v, exists := f.Value(); !exists {
		panic("value not set")
	} else {
		return v
	}
}

func (f *OpFloat) ValueDefault(defaultValue float64) float64 {
	if v, exists := f.Value(); !exists {
		return defaultValue
	} else {
		return v
	}
}

func (f *OpFloat) Set(value float64) {
	if !f.Exists() {
		f.value = &f.store
	}
	f.store = value
}

func (f *OpFloat) Clear() {
	f.value = nil
}

func (f OpFloat) MarshalJSON() ([]byte, error) {
	if val, exists := f.Value(); exists {
		return []byte(fmt.Sprint(val)), nil
	} else {
		return []byte("null"), nil
	}
}

func (f *OpFloat) UnmarshalJSON(bytes []byte) error {
	val, err := strconv.ParseFloat(string(bytes), 64)
	if err != nil {
		return err
	}
	f.Set(val)
	return nil
}
