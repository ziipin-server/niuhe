package niuhe

import (
	"github.com/gin-gonic/gin"
	"github.com/ziipin-server/zpform"
	"reflect"
	"regexp"
	"strings"
)

const (
	GET      int = 1
	POST     int = 2
	GET_POST int = 3
)

type routeInfo struct {
	Methods    int
	Path       string
	handleFunc gin.HandlerFunc
}

type Module struct {
	urlPrefix string
	handlers  []routeInfo
}

func NewModule(urlPrefix string) *Module {
	return &Module{
		urlPrefix: urlPrefix,
		handlers:  make([]routeInfo, 0),
	}
}

type Context struct {
	gin.Context
}

func newContext(c *gin.Context) *Context {
	return &Context{*c}
}

func parseName(camelName string) string {
	re := regexp.MustCompile("[A-Z][a-z0-9]*")
	parts := re.FindAllString(camelName, -1)
	return strings.Join(parts, "_")
}

func (mod *Module) Register(group interface{}) *Module {
	groupType := reflect.TypeOf(group)
	groupName := groupType.Elem().Name()
	for i := 0; i < groupType.NumMethod(); i++ {
		m := groupType.Method(i)
		name := m.Name
		var methods int
		if strings.HasSuffix(name, "_GET") {
			methods = GET
			name = name[:len(name)-len("_GET")]
		} else if strings.HasSuffix(name, "_POST") {
			methods = POST
			name = name[:len(name)-len("_POST")]
		} else {
			methods = GET_POST
		}
		path := strings.ToLower("/" + parseName(groupName) + "/" + parseName(name) + "/")
		mod._Register(methods, path, m.Func)
	}
	return mod
}

var bindFunc reflect.Value

func (mod *Module) _Register(methods int, path string, funcValue reflect.Value) *Module {
	funcType := funcValue.Type()
	if funcType.Kind() != reflect.Func {
		panic("handleFunc必须为函数")
	}
	if funcType.NumIn() != 4 || funcType.NumOut() != 1 {
		panic("handleFunc必须有三个参数,并且只返回一个error")
	}
	groupType := funcType.In(0)
	nilGroupValue := reflect.Zero(groupType)
	reqType := funcType.In(2).Elem()
	rspType := funcType.In(3).Elem()
	ginHandler := func(c *gin.Context) {
		req := reflect.New(reqType)
		rsp := reflect.New(rspType)
		var ierr interface{}
		if formErr := zpform.ReadReflectedStructForm(c.Request, req); formErr != nil {
			ierr = formErr
		}
		if ierr == nil {
			outs := funcValue.Call([]reflect.Value{
				nilGroupValue,
				reflect.ValueOf(newContext(c)),
				req,
				rsp,
			})
			ierr = outs[0].Interface()
		}
		if ierr != nil {
			commErr, ok := ierr.(ICommError)
			if ok {
				c.JSON(200, map[string]interface{}{
					"result":  commErr.GetCode(),
					"message": commErr.GetMessage(),
				})
			} else if err, ok := ierr.(error); ok {
				c.JSON(200, map[string]interface{}{
					"result":  -1,
					"message": err.Error(),
				})
			} else {
				c.JSON(200, map[string]interface{}{
					"result":  -1,
					"message": "Unknown",
				})
			}
		} else {
			c.JSON(200, map[string]interface{}{
				"result": 0,
				"data":   rsp.Interface(),
			})
		}
	}
	mod.handlers = append(mod.handlers, routeInfo{Methods: methods, Path: path, handleFunc: ginHandler})
	return mod
}

func init() {
	cType := reflect.TypeOf(&gin.Context{})
	bindMethod, found := cType.MethodByName("Bind")
	if !found {
		panic("Cannot find Bind mathod")
	}
	bindFunc = bindMethod.Func
}
