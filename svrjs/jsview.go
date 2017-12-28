package svrjs

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/robertkrimen/otto"
	"github.com/ziipin-server/niuhe"
)

type JsView struct {
	BasePath    string
	ScriptEntry string
	ModulePath  []string
	env         *Env
	script      *otto.Script
}

func (view *JsView) Init(basePath, scriptEntry string, modulePath []string) {
	view.BasePath = basePath
	view.ScriptEntry = scriptEntry
	view.ModulePath = modulePath
	view.env = NewEnv(modulePath)
	var err error
	view.script, err = view.env.vm.Compile(scriptEntry, nil)
	if err != nil {
		panic(err)
	}
}

func (view *JsView) Register(mod *niuhe.Module, pf niuhe.IApiProtocolFactory,
	middlewares ...niuhe.HandlerFunc) {
	gv := reflect.ValueOf(view)
	gt := reflect.TypeOf(gv).Elem()
	f := gt.MethodByName("HandleRequest")
	mod.AddCustomRoute(
		niuhe.GET_POST,
		fmt.Sprintf("/%s/", view.BasePath),
		reflect.ValueOf(view),
		f.Func,
		pf,
		middlewares,
	)
}

func (view *JsView) HandleRequest(c *niuhe.Context) {
	env.vm.Set("c", c)
	if _, err := env.Run(view.script); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}
