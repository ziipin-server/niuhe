package niuhe

import (
	"fmt"
	"strconv"
)

type OpLong struct {
	value *int64
	store int64
}

func (f *OpLong) Exists() bool {
	return f.value != nil
}

func (f *OpLong) Value() (int64, bool) {
	if !f.Exists() {
		return 0, false
	}
	return *f.value, true
}

func (f *OpLong) MustValue() int64 {
	if v, exists := f.Value(); !exists {
		panic("value not set")
	} else {
		return v
	}
}

func (f *OpLong) ValueDefault(defaultValue int64) int64 {
	if v, exists := f.Value(); !exists {
		return defaultValue
	} else {
		return v
	}
}

func (f *OpLong) Set(value int64) {
	if !f.Exists() {
		f.value = &f.store
	}
	f.store = value
}

func (f *OpLong) Clear() {
	f.value = nil
}

func (f OpLong) MarshalJSON() ([]byte, error) {
	if val, exists := f.Value(); exists {
		return []byte(fmt.Sprint(val)), nil
	} else {
		return []byte("null"), nil
	}
}

func (f *OpLong) UnmarshalJSON(bytes []byte) error {
	val, err := strconv.ParseInt(string(bytes), 10, 64)
	if err != nil {
		return err
	}
	f.Set(val)
	return nil
}
