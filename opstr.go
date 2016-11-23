package niuhe

import "encoding/json"

type OpStr struct {
	value *string
	store string
}

func (f *OpStr) Exists() bool {
	return f.value != nil
}

func (f *OpStr) Value() (string, bool) {
	if !f.Exists() {
		return "", false
	}
	return *f.value, true
}

func (f *OpStr) MustValue() string {
	if v, exists := f.Value(); !exists {
		panic("value not set")
	} else {
		return v
	}
}

func (f *OpStr) ValueDefault(defaultValue string) string {
	if v, exists := f.Value(); !exists {
		return defaultValue
	} else {
		return v
	}
}

func (f *OpStr) Set(value string) {
	if !f.Exists() {
		f.value = &f.store
	}
	f.store = value
}

func (f *OpStr) Clear() {
	f.value = nil
}

func (f OpStr) MarshalJSON() ([]byte, error) {
	if val, exists := f.Value(); exists {
		return json.Marshal(val)
	} else {
		return []byte("null"), nil
	}
}

func (f *OpStr) UnmarshalJSON(bytes []byte) error {
	var val string
	if err := json.Unmarshal(bytes, &val); err != nil {
		return err
	}
	f.Set(val)
	return nil
}
