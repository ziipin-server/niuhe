package orm

import (
	"reflect"

	"github.com/robertkrimen/otto"

	"github.com/go-xorm/xorm"
)

type model struct {
	engine     **xorm.Engine
	singleType reflect.Type
	sliceType  reflect.Type
}

func NewModel(m interface{}, engine **xorm.Engine) *model {
	t := reflect.TypeOf(m)
	return &model{
		engine:     engine,
		singleType: t.Elem(),
		sliceType:  reflect.SliceOf(t),
	}
}

//New create a new pointer that pointing to a model struct
func (m *model) New(fieldValues ...map[string]interface{}) interface{} {
	pv := reflect.New(m.singleType)
	if len(fieldValues) > 0 {
		v := pv.Elem()
		for name, value := range fieldValues[0] {
			vv := reflect.ValueOf(value)
			field := v.FieldByName(name)
			if vv.Kind() == reflect.Float64 {
				switch field.Kind() {
				case reflect.Int:
					fallthrough
				case reflect.Int8:
					fallthrough
				case reflect.Int16:
					fallthrough
				case reflect.Int32:
					fallthrough
				case reflect.Int64:
					{
						fv := vv.Float()
						field.SetInt(int64(fv))
					}
				case reflect.Uint:
					fallthrough
				case reflect.Uint8:
					fallthrough
				case reflect.Uint16:
					fallthrough
				case reflect.Uint32:
					fallthrough
				case reflect.Uint64:
					{
						fv := vv.Float()
						field.SetUint(uint64(fv))
					}
				case reflect.Float32:
					field.SetFloat(vv.Float())
				default:
					field.Set(vv)
				}
			} else {
				field.Set(vv)
			}
		}
	}
	return pv.Interface()
}

//NewSlice create a new pointer that pointing to a model slice
func (m *model) NewSlice() interface{} {
	return reflect.New(m.sliceType).Interface()
}

func (m *model) Where(query interface{}, args ...interface{}) *session {
	return newSession(m, true).And(query, args...)
}

func (m *model) ID(id ...interface{}) *session {
	return newSession(m, true).ID(id...)
}

func (m *model) In(column string, args ...interface{}) *session {
	return newSession(m, true).In(column, args...)
}

func (m *model) NotIn(column string, args ...interface{}) *session {
	return newSession(m, true).NotIn(column, args...)
}

func (m *model) Limit(limit int, start ...int) *session {
	return newSession(m, true).Limit(limit, start...)
}

func (m *model) Asc(cols ...string) *session {
	return newSession(m, true).Asc(cols...)
}

func (m *model) Desc(cols ...string) *session {
	return newSession(m, true).Desc(cols...)
}

func (m *model) Get(call otto.FunctionCall) otto.Value {
	return newSession(m, true).Get(call)
}

func (m *model) Find(call otto.FunctionCall) otto.Value {
	return newSession(m, true).Find(call)
}

func (m *model) Session(call otto.FunctionCall) otto.Value {
	callback := call.Argument(0)
	vm := call.Otto
	if !callback.IsFunction() {
		panic(vm.MakeTypeError("callback must be a function"))
	}
	s := newSession(m, false)
	callback.Call(otto.NullValue(), s)
	s.xs.Close()
	return otto.UndefinedValue()
}

func (m *model) Atom(call otto.FunctionCall) otto.Value {
	callback := call.Argument(0)
	vm := call.Otto
	if !callback.IsFunction() {
		panic(vm.MakeTypeError("callback must be a function"))
	}
	s := newSession(m, false)
	if err := s.xs.Begin(); err != nil {
		panic(vm.MakeCustomError("DBError", "begin transaction fail: "+err.Error()))
	}
	ret, err := callback.Call(otto.NullValue(), s)
	shouldRollback := err != nil
	if err == nil {
		if ret.IsBoolean() {
			bv, _ := ret.ToBoolean()
			shouldRollback = !bv
		}
	}
	if shouldRollback {
		if err := s.xs.Rollback(); err != nil {
			panic(vm.MakeCustomError("DBError", "rollback transaction fail: "+err.Error()))
		}
	} else {
		if err := s.xs.Commit(); err != nil {
			panic(vm.MakeCustomError("DBError", "commit transaction fail: "+err.Error()))
		}
	}
	s.xs.Close()
	if err != nil {
		panic(err)
	}
	return otto.UndefinedValue()
}

func (m *model) WithSession(parent *session) *session {
	return attachSession(m, parent)
}

func GetModelsModule(defaultEngine **xorm.Engine, modelInstances ...interface{}) (string, interface{}) {
	modelsMap := make(map[string]interface{}, len(modelInstances))
	for _, mi := range modelInstances {
		var targetModel *model
		if m, ok := mi.(*model); ok {
			targetModel = m
		} else {
			targetModel = NewModel(mi, defaultEngine)
		}
		modelsMap[targetModel.singleType.Name()] = targetModel
	}
	return "models", modelsMap
}
