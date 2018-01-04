package svrjs

import (
	"fmt"
	"net/http"
	"os"
	"reflect"

	"github.com/robertkrimen/otto"
	"github.com/ziipin-server/niuhe"
)

type JsView struct {
	BasePath    string
	ScriptEntry string
	ModulePath  []string
	debugging   bool
	entryInfo   os.FileInfo
	loader      *ScriptLoader
	script      *otto.Script
	builtins    map[string]interface{}
}

func NewJsView(basePath, scriptEntry string, modulePath []string, debugging bool) *JsView {
	view := &JsView{
		BasePath:    basePath,
		ScriptEntry: scriptEntry,
		ModulePath:  modulePath,
		debugging:   debugging,
		loader:      NewScriptLoader(modulePath, debugging),
		builtins:    make(map[string]interface{}),
	}
	return view
}

func (view *JsView) Register(mod *niuhe.Module, pf niuhe.IApiProtocolFactory, middlewares ...niuhe.HandlerFunc) {
	gt := reflect.TypeOf(view)
	f, found := gt.MethodByName("HandleRequest")
	if !found {
		panic("Cannot find method HandleRequest in type " + gt.Name())
	}
	mod.AddCustomRoute(
		niuhe.GET_POST,
		fmt.Sprintf("/%s/*path", view.BasePath),
		reflect.ValueOf(view),
		f.Func,
		pf,
		middlewares,
	)
}

func (view *JsView) newEnv(additionalModules map[string]interface{}) *Env {
	env := NewEnv(view.loader, view.debugging)
	for id, content := range view.builtins {
		env.InstallBuiltinModule(id, content)
	}
	for id, content := range additionalModules {
		env.InstallBuiltinModule(id, content)
	}
	return env
}

func (view *JsView) releaseEnv(env *Env) {
}

func (view *JsView) HandleRequest(c *niuhe.Context) {
	path := c.Param("path")
	env := view.newEnv(map[string]interface{}{
		"niuhe-context": c,
		"routePath":     path,
	})
	defer view.releaseEnv(env)
	if view.script == nil || view.debugging {
		var err error
		var newInfo os.FileInfo
		if newInfo, err = os.Stat(view.ScriptEntry); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		} else if view.script == nil || newInfo.ModTime().After(view.entryInfo.ModTime()) {
			if view.script, err = env.vm.Compile(view.ScriptEntry, nil); err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			view.entryInfo = newInfo
		}
	}
	if _, err := env.Run(view.script); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}

func (view *JsView) InstallModule(id string, content interface{}) *JsView {
	view.builtins[id] = content
	return view
}
