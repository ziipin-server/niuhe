package orm

import (
	"github.com/go-xorm/xorm"
	"github.com/ziipin-server/niuhe/svrjs"
)

type EngineWrapper struct {
	*xorm.Engine
}

func EngineFactory(pe **xorm.Engine) svrjs.BuiltinModuleFactory {
	return svrjs.MakeFactory(func() interface{} {
		return &EngineWrapper{*pe}
	})
}

func (ew *EngineWrapper) Query(sqlOrArgs ...interface{}) []map[string]string {
	results, err := ew.Engine.Query(sqlOrArgs...)
	if err != nil {
		panic(err)
	}
	r := make([]map[string]string, len(results))
	for i, row := range results {
		r[i] = make(map[string]string, len(row))
		for k, v := range row {
			r[i][k] = string(v)
		}
	}
	return r
}
