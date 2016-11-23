package niuhe

import "strconv"

type OpInt struct {
	value *int
	store int
}

func (f *OpInt) Exists() bool {
	return f.value != nil
}

func (f *OpInt) Value() (int, bool) {
	if !f.Exists() {
		return 0, false
	}
	return *f.value, true
}

func (f *OpInt) MustValue() int {
	if v, exists := f.Value(); !exists {
		panic("value not set")
	} else {
		return v
	}
}

func (f *OpInt) ValueDefault(defaultValue int) int {
	if v, exists := f.Value(); !exists {
		return defaultValue
	} else {
		return v
	}
}

func (f *OpInt) Set(value int) {
	if !f.Exists() {
		f.value = &f.store
	}
	f.store = value
}

func (f *OpInt) Clear() {
	f.value = nil
}

func (f OpInt) MarshalJSON() ([]byte, error) {
	if val, exists := f.Value(); exists {
		return []byte(strconv.Itoa(val)), nil
	} else {
		return []byte("null"), nil
	}
}

func (f *OpInt) UnmarshalJSON(bytes []byte) error {
	val, err := strconv.Atoi(string(bytes))
	if err != nil {
		return err
	}
	f.Set(val)
	return nil
}
