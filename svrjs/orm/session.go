package orm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/robertkrimen/otto"
)

type session struct {
	model     *model
	xs        *xorm.Session
	autoclose bool
}

func newSession(model *model, autoclose bool) *session {
	s := &session{
		model:     model,
		xs:        (*(model.engine)).NewSession(),
		autoclose: autoclose,
	}
	return s
}

func attachSession(model *model, parent *session) *session {
	if parent.model.engine != model.engine {
		panic("cannot attach to the model with different engine")
	}
	return &session{
		model:     model,
		xs:        parent.xs,
		autoclose: false,
	}
}

func mustInt(vm *otto.Otto, v otto.Value, name string) int {
	i64, err := v.ToInteger()
	if err != nil {
		panic(vm.MakeTypeError(fmt.Sprintf("%s must be an integer", name)))
	}
	return int(i64)
}

func mustStr(vm *otto.Otto, v otto.Value, name string) string {
	s, err := v.ToString()
	if err != nil {
		panic(vm.MakeTypeError(fmt.Sprintf("%s must be an integer", name)))
	}
	return s
}

func (s *session) handleCondition(vm *otto.Otto, args []otto.Value) {
	if len(args) == 1 && args[0].IsObject() {
		argObj := args[0].Object()
		for _, k := range argObj.Keys() {
			var op, realKey string
			if parts := strings.SplitN(k, "__", 2); len(parts) > 1 {
				realKey = parts[0]
				switch strings.ToLower(parts[1]) {
				case "eq":
					op = "= ?"
				case "ne":
					op = "<> ?"
				case "lt":
					op = "< ?"
				case "gt":
					op = "> ?"
				case "le":
					op = "<= ?"
				case "ge":
					op = ">= ?"
				}
			} else {
				realKey = k
				op = "= ?"
			}
			v, _ := argObj.Get(k)
			vi, _ := v.Export()
			s.xs = s.xs.And(realKey+op, vi)
		}
	} else if len(args) > 0 {
		pks := make(core.PK, len(args))
		var err error
		for i, arg := range args {
			pks[i], err = arg.Export()
			if err != nil {
				panic(vm.MakeCustomError("orm get", "parse pk fail: "+err.Error()))
			}
		}
		s.xs = s.xs.Id(pks)
	}
}

func (s *session) ID(id ...interface{}) *session {
	s.xs = s.xs.ID(core.PK(id))
	return s
}

func (s *session) And(query interface{}, args ...interface{}) *session {
	s.xs = s.xs.And(query, args...)
	return s
}

func (s *session) Or(query interface{}, args ...interface{}) *session {
	s.xs = s.xs.Or(query, args...)
	return s
}

func (s *session) In(column string, args ...interface{}) *session {
	s.xs = s.xs.In(column, args...)
	return s
}

func (s *session) NotIn(column string, args ...interface{}) *session {
	s.xs = s.xs.NotIn(column, args...)
	return s
}

func (s *session) Limit(limit int, start ...int) *session {
	s.xs = s.xs.Limit(limit, start...)
	return s
}

func (s *session) Asc(cols ...string) *session {
	s.xs = s.xs.Asc(cols...)
	return s
}

func (s *session) Desc(cols ...string) *session {
	s.xs = s.xs.Desc(cols...)
	return s
}

func (s *session) Close() {
	s.xs.Close()
}

func (s *session) Get(call otto.FunctionCall) otto.Value {
	if s.autoclose {
		defer s.Close()
	}
	args := call.ArgumentList
	vm := call.Otto
	s.handleCondition(vm, args)
	m := s.model.New()
	has, err := s.xs.Get(m)
	if err != nil {
		panic(vm.MakeCustomError("orm get", "db error: "+err.Error()))
	}
	if has {
		v, _ := vm.ToValue(m)
		return v
	}
	return otto.NullValue()
}

func (s *session) Find(call otto.FunctionCall) otto.Value {
	if s.autoclose {
		defer s.Close()
	}
	args := call.ArgumentList
	vm := call.Otto
	s.handleCondition(vm, args)
	l := s.model.NewSlice()
	if err := s.xs.Find(l); err != nil {
		panic(vm.MakeCustomError("orm find", "db error: "+err.Error()))
	}
	v, _ := vm.ToValue(reflect.ValueOf(l).Elem().Interface())
	return v
}

func (s *session) Count(call otto.FunctionCall) otto.Value {
	if s.autoclose {
		defer s.Close()
	}
	args := call.ArgumentList
	vm := call.Otto
	s.handleCondition(vm, args)
	bean := s.model.New()
	if count, err := s.xs.Count(bean); err != nil {
		panic(vm.MakeCustomError("orm count", "db error: "+err.Error()))
	} else {
		v, _ := vm.ToValue(count)
		return v
	}
}

func (s *session) Update(call otto.FunctionCall) otto.Value {
	if s.autoclose {
		defer s.Close()
	}
	vm := call.Otto
	arg := call.Argument(0)
	if !arg.IsObject() {
		panic(vm.MakeTypeError("Update argument must be an object"))
	}
	newValues := arg.Object()
	nv := make(map[string]interface{})
	for _, name := range newValues.Keys() {
		v, err := newValues.Get(name)
		if err != nil {
			panic(vm.MakeCustomError("ValueError", "cannot read field "+name))
		}
		vi, err := v.Export()
		if err != nil {
			panic(vm.MakeCustomError("ValueError", "cannot export field "+name))
		}
		nv[name] = vi
	}
	affected, err := s.xs.Table(s.model.New()).Update(nv)
	if err != nil {
		panic(vm.MakeCustomError("DBError", "update error: "+err.Error()))
	}
	ret, _ := vm.ToValue(affected)
	return ret
}
