package svrjs

import (
	"path"

	"github.com/robertkrimen/otto"
)

type BuiltinModuleFactory interface {
	CreateModule(*otto.Otto) otto.Value
}

type simpleFactory struct {
	fn func() interface{}
}

func (f *simpleFactory) CreateModule(vm *otto.Otto) otto.Value {
	val, err := vm.ToValue(f.fn())
	if err != nil {
		panic(err)
	}
	return val
}

func MakeFactory(fn func() interface{}) BuiltinModuleFactory {
	return &simpleFactory{fn: fn}
}

//Env env struct
type Env struct {
	debugging              bool
	vm                     *otto.Otto
	loader                 *ScriptLoader
	builtinModules         map[string]otto.Value
	builtinModuleFactories map[string]BuiltinModuleFactory
}

const (
	debugRequireSrc = `
		var require = (function() {
			var __getRequire = function(cwd) {
				return function(id) {
					var modPath = __getModulePath(cwd, id)
					var loaded = __loadSource(modPath)
					if (loaded.isBuiltin) {
						return loaded.builtin
					}
					var module = {exports:{}}
					loaded.src(
						module.exports,
						__getRequire(loaded.dirname),
						module,
						loaded.filename,
						loaded.dirname
					)
					return module.exports
				}
			}
			return __getRequire('./')
		})()
	`
	releaseRequireSrc = `
		var require = (function() {
			var __modules = {}
			var __getRequire = function(cwd) {
				return function(id) {
					var modPath = __getModulePath(cwd, id)
					if (!__modules[modPath]) {
						var loaded = __loadSource(modPath)
						if (loaded.isBuiltin) {
							return loaded.builtin
						}
						var module = {exports:{}}
						loaded.src(
							module.exports,
							__getRequire(loaded.dirname),
							module,
							loaded.filename,
							loaded.dirname
						)
						__modules[modPath] = module.exports
					}
					return __modules[modPath]
				}
			}
			return __getRequire('./')
		})()
	`
)

//NewEnv create a new environment
func NewEnv(loader *ScriptLoader, debugging bool) *Env {
	env := &Env{
		vm:                     otto.New(),
		loader:                 loader,
		debugging:              debugging,
		builtinModules:         make(map[string]otto.Value),
		builtinModuleFactories: make(map[string]BuiltinModuleFactory),
	}
	env.vm.Set("__getModulePath", func(call otto.FunctionCall) otto.Value {
		vm := call.Otto
		var cwd, moduleID string
		var err error
		if cwd, err = call.Argument(0).ToString(); err != nil {
			panic(vm.MakeCustomError("__getModulePath error", err.Error()))
		}
		if moduleID, err = call.Argument(1).ToString(); err != nil {
			panic(vm.MakeCustomError("__getModulePath error", err.Error()))
		}
		if _, found := env.builtinModules[moduleID]; found {
			ret, _ := otto.ToValue(moduleID)
			return ret
		}
		if _, found := env.builtinModuleFactories[moduleID]; found {
			ret, _ := otto.ToValue(moduleID)
			return ret
		}
		if ap, err := env.loader.GetModuleAbs(cwd, moduleID); err != nil {
			panic(vm.MakeCustomError("__getModulePath error", err.Error()))
		} else {
			ret, _ := otto.ToValue(ap)
			return ret
		}
	})
	var requireSrc string
	if env.debugging {
		requireSrc = debugRequireSrc
	} else {
		requireSrc = releaseRequireSrc
	}
	env.vm.Set("__loadSource", func(call otto.FunctionCall) otto.Value {
		var mp string
		var err error
		vm := call.Otto
		// reading arguments
		if mp, err = call.Argument(0).ToString(); err != nil {
			panic(vm.MakeCustomError("__loadSource error", err.Error()))
		}
		// finding built builtin modules
		if mod, found := env.builtinModules[mp]; found {
			retObj, _ := vm.Object("({isBuiltin: true})")
			retObj.Set("builtin", mod)
			return retObj.Value()
		}
		// finding unbuilt builtin modules
		if mf, found := env.builtinModuleFactories[mp]; found {
			retObj, _ := vm.Object("({isBuiltin: true})")
			mod := mf.CreateModule(vm)
			retObj.Set("builtin", mod)
			env.builtinModules[mp] = mod
			return retObj.Value()
		}
		// loading module on file system
		src, err := env.loader.LoadScript(mp)
		if err != nil {
			panic(vm.MakeCustomError("__loadSource error", err.Error()))
		}
		script, err := vm.Compile(mp, src)
		if err != nil {
			panic(vm.MakeCustomError("__loadSource error", err.Error()))
		}
		modValue, err := vm.Run(script)
		if err != nil {
			panic(vm.MakeCustomError("__loadSource error", err.Error()))
		}
		retObj, _ := vm.Object("({})")
		retObj.Set("src", modValue)
		retObj.Set("filename", path.Base(mp))
		retObj.Set("dirname", path.Dir(mp))
		return retObj.Value()
	})
	_, err := env.vm.Run(requireSrc)
	if err != nil {
		switch err.(type) {
		case *otto.Error:
			panic(err.(otto.Error).String())
		default:
			panic(err)
		}
	}
	return env
}

//Run run a code object, may be string, []byte, *otto.Script, etc
func (env *Env) Run(code interface{}) (otto.Value, error) {
	return env.vm.Run(code)
}

//InstallBuiltinModule just like its name
func (env *Env) InstallBuiltinModule(id string, content interface{}) {
	val, err := env.vm.ToValue(content)
	if err != nil {
		panic(err)
	}
	if mf, ok := content.(BuiltinModuleFactory); ok {
		env.builtinModuleFactories[id] = mf
	} else {
		env.builtinModules[id] = val
	}
}
