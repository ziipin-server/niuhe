package niuhe

import "strconv"

type OpBool struct {
	value *bool
	store bool
}

func (f *OpBool) Exists() bool {
	return f.value != nil
}

func (f *OpBool) Value() (bool, bool) {
	if !f.Exists() {
		return false, false
	}
	return *f.value, true
}

func (f *OpBool) MustValue() bool {
	if v, exists := f.Value(); !exists {
		panic("value not set")
	} else {
		return v
	}
}

func (f *OpBool) ValueDefault(defaultValue bool) bool {
	if v, exists := f.Value(); !exists {
		return defaultValue
	} else {
		return v
	}
}

func (f *OpBool) Set(value bool) {
	if !f.Exists() {
		f.value = &f.store
	}
	f.store = value
}

func (f *OpBool) Clear() {
	f.value = nil
}

func (f OpBool) MarshalJSON() ([]byte, error) {
	if val, exists := f.Value(); exists {
		return []byte(strconv.FormatBool(val)), nil
	} else {
		return []byte("null"), nil
	}
}

func (f *OpBool) UnmarshalJSON(bytes []byte) error {
	val, err := strconv.ParseBool(string(bytes))
	if err != nil {
		return err
	}
	f.Set(val)
	return nil
}
