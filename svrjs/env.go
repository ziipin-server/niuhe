package svrjs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/robertkrimen/otto"
)

//Env env struct
type Env struct {
	vm             *otto.Otto
	builtinModules map[string]otto.Value
}

//NewEnv create a new environment
func NewEnv(searchPaths []string) *Env {
	env := &Env{}
	env.vm = otto.New()
	if len(searchPaths) == 0 {
		searchPaths = []string{"./"}
	}
	env.builtinModules = make(map[string]otto.Value)
	var checkFile = func(fp string) (string, bool) {
		fp, _ = filepath.Abs(fp)
		fi, err := os.Stat(fp)
		return fp, !os.IsNotExist(err) && !fi.IsDir()
	}
	env.vm.Set("__getModulePath", func(call otto.FunctionCall) otto.Value {
		vm := call.Otto
		var err error
		var cwd, moduleID string
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
		var realPaths []string
		if strings.HasPrefix(moduleID, ".") {
			realPaths = []string{cwd}
		} else {
			realPaths = searchPaths
		}
		for _, sp := range realPaths {
			mp := path.Join(sp, moduleID)
			if !strings.HasSuffix(mp, ".js") {
				if fn, found := checkFile(mp + ".js"); found {
					ret, _ := otto.ToValue(fn)
					return ret
				}
				if fn, found := checkFile(path.Join(mp, "index.js")); found {
					ret, _ := otto.ToValue(fn)
					return ret
				}
			}
		}
		panic(vm.MakeCustomError("__getModulePath error", fmt.Sprintf("%s not found", moduleID)))
	})
	env.vm.Set("__loadSource", func(call otto.FunctionCall) otto.Value {
		var mp string
		var err error
		vm := call.Otto
		if mp, err = call.Argument(0).ToString(); err != nil {
			panic(vm.MakeCustomError("__loadSource error", err.Error()))
		}
		if mod, found := env.builtinModules[mp]; found {
			retObj, _ := vm.Object("({isBuiltin: true})")
			retObj.Set("builtin", mod)
			return retObj.Value()
		}
		srcBytes, err := ioutil.ReadFile(mp)
		if err != nil {
			panic(vm.MakeCustomError("__loadSource error", fmt.Sprintf("load module %s fail: %s", mp, err.Error())))
		}
		dirname := path.Dir(mp)
		filename := path.Base(mp)
		src := "(function(exports, require, module, __filename, __dirname){\n" + string(srcBytes) + "\n})"
		retObj, _ := vm.Object("({})")
		retObj.Set("src", src)
		retObj.Set("filename", filename)
		retObj.Set("dirname", dirname)
		return retObj.Value()
	})
	_, err := env.vm.Run(`
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
						var src = loaded.src
						eval(src)(
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
	`)
	if err != nil {
		switch err.(type) {
		case otto.Error:
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
	env.builtinModules[id] = val
}
