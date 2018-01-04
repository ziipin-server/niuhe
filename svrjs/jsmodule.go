package svrjs

import "github.com/ziipin-server/niuhe"

//NewJsModule create a js module
func NewJsModule(urlPrefix, scriptEntry string, paths []string, debugging bool) (*niuhe.Module, *JsView) {
	m := niuhe.NewModule(urlPrefix)
	view := NewJsView("", scriptEntry, paths, debugging)
	m.Register(view)
	return m, view
}
