package svrjs

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/robertkrimen/otto"
	"github.com/ziipin-server/niuhe"
)

func getContext(vm *otto.Otto) *niuhe.Context {
	var c *niuhe.Context
	var ok bool
	if cv, err := vm.Run("(require('niuhe-context'))"); err != nil {
		panic(vm.MakeCustomError("ValueError", "get context fail: "+err.Error()))
	} else if ci, err := cv.Export(); err != nil {
		panic(vm.MakeCustomError("ValueError", "get context fail: "+err.Error()))
	} else if c, ok = ci.(*niuhe.Context); !ok {
		panic(vm.MakeCustomError("ValueError", "get context fail: c type error"))
	}
	return c
}

func getRenderStatusCode(vm *otto.Otto, args []otto.Value) (int, []otto.Value) {
	if len(args) < 1 {
		panic(vm.MakeTypeError("arguments length must >= 1"))
	}
	if args[0].IsNumber() {
		code, _ := args[0].ToInteger()
		return int(code), args[1:]
	}
	return 200, args
}

func GetWebModule() (string, interface{}) {
	return "web", map[string]interface{}{
		"renderText": func(call otto.FunctionCall) otto.Value {
			vm := call.Otto
			c := getContext(vm)
			statusCode, args := getRenderStatusCode(vm, call.ArgumentList)
			var err error
			var content string
			if len(args) < 1 {
				panic(vm.MakeTypeError(call.CallerLocation() + ": Not enough arguments"))
			}
			content, err = args[0].ToString()
			contentArgs := make([]interface{}, len(args)-1)
			for i := 1; i < len(args); i++ {
				contentArgs[i-1], err = args[i].Export()
				if err != nil {
					panic(vm.MakeTypeError(call.CallerLocation() + ": enough error " + err.Error()))
				}
			}
			c.String(statusCode, content, contentArgs...)
			return otto.UndefinedValue()
		},
		"renderHTML": func(call otto.FunctionCall) otto.Value {
			vm := call.Otto
			c := getContext(vm)
			statusCode, args := getRenderStatusCode(vm, call.ArgumentList)
			templateName, err := args[0].ToString()
			if err != nil {
				panic(vm.MakeTypeError("argument[1] must be a string"))
			}
			templateArgs, err := args[1].Export()
			if err != nil {
				panic(vm.MakeTypeError(call.CallerLocation() + ": template arguments export fail"))
			}
			if ta, ok := templateArgs.(map[string]interface{}); !ok {
				panic(vm.MakeTypeError(call.CallerLocation() + ": template arguments must be a map"))
			} else {
				c.HTML(int(statusCode), templateName, pongo2.Context(ta))
			}
			return otto.UndefinedValue()
		},
		"renderJSON": func(call otto.FunctionCall) otto.Value {
			vm := call.Otto
			c := getContext(vm)
			statusCode, args := getRenderStatusCode(vm, call.ArgumentList)
			if len(args) != 1 {
				panic(vm.MakeTypeError(call.CallerLocation() + ": Not enough arguments"))
			}
			obj, err := args[0].Export()
			if err != nil {
				panic(vm.MakeTypeError(call.CallerLocation() + ": argument export fail " + err.Error()))
			}
			c.JSON(statusCode, obj)
			return otto.UndefinedValue()
		},
		"$GET": func(call otto.FunctionCall) otto.Value {
			vm := call.Otto
			c := getContext(vm)
			args := call.ArgumentList
			if len(args) != 1 {
				panic(vm.MakeTypeError("arguments length must be 1"))
			}
			key, err := args[0].ToString()
			if err != nil {
				panic(vm.MakeTypeError("argument[0] must be a string"))
			}
			if v, exists := c.GetQuery(key); exists {
				sv, _ := vm.ToValue(v)
				return sv
			}
			return otto.NullValue()
		},
		"$POST": func(call otto.FunctionCall) otto.Value {
			vm := call.Otto
			c := getContext(vm)
			args := call.ArgumentList
			if len(args) != 1 {
				panic(vm.MakeTypeError("arguments length must be 1"))
			}
			key, err := args[0].ToString()
			if err != nil {
				panic(vm.MakeTypeError("argument[0] must be a string"))
			}
			if v, exists := c.GetPostForm(key); exists {
				sv, _ := vm.ToValue(v)
				return sv
			}
			return otto.NullValue()
		},
		"$METHOD": func(call otto.FunctionCall) otto.Value {
			vm := call.Otto
			c := getContext(vm)
			r, _ := vm.ToValue(strings.ToUpper(c.Request.Method))
			return r
		},
		"$REMOTE_ADDR": func(call otto.FunctionCall) otto.Value {
			vm := call.Otto
			c := getContext(vm)
			r, _ := vm.ToValue(c.ClientIP())
			return r
		},
		"$SESSION": func(call otto.FunctionCall) otto.Value {
			vm := call.Otto
			c := getContext(vm)
			args := call.ArgumentList
			switch len(args) {
			case 1:
				{
					skey, err := args[0].ToString()
					if err != nil {
						panic(vm.MakeTypeError(call.CallerLocation() + ":arguments error"))
					}
					sval := c.GetSession(skey)
					if sval == nil {
						return otto.NullValue()
					}
					jsval, err := vm.ToValue(sval)
					if err != nil {
						panic(vm.MakeTypeError(call.CallerLocation() + ":session value error " + err.Error()))
					}
					return jsval
				}
			case 2:
				{
					skey, err := args[0].ToString()
					if err != nil {
						panic(vm.MakeTypeError(call.CallerLocation() + ":arguments error"))
					}
					if args[1].IsNull() {
						c.DelSession(skey)
					} else {
						sval, err := args[1].Export()
						if err != nil {
							panic(vm.MakeTypeError(call.CallerLocation() + ":arguments error"))
						}
						c.SetSession(skey, sval)
					}
					return otto.UndefinedValue()
				}
			default:
				panic(vm.MakeTypeError(call.CallerLocation() + ":arguments error"))
			}
		},
		"$URL": func(call otto.FunctionCall) otto.Value {
			vm := call.Otto
			c := getContext(vm)
			u, _ := vm.ToValue(c.Request.URL)
			return u
		},
		"Redirect": func(call otto.FunctionCall) otto.Value {
			vm := call.Otto
			c := getContext(vm)
			target := call.Argument(0)
			if !target.IsString() {
				panic(vm.MakeTypeError(call.CallerLocation() + ":arguments error"))
			}
			location, _ := target.ToString()
			c.Redirect(302, location)
			return otto.UndefinedValue()
		},
	}
}

func GetUtilsModule() (string, interface{}) {
	return "utils", map[string]interface{}{
		"md5": func(src interface{}) string {
			var srcBytes []byte
			switch src.(type) {
			case []byte:
				srcBytes = src.([]byte)
			case string:
				srcBytes = []byte(src.(string))
			default:
				srcBytes = []byte(fmt.Sprint(src))
			}
			md5Result := md5.Sum(srcBytes)
			return hex.EncodeToString(md5Result[:])
		},
		"MD5": func(src interface{}) string {
			var srcBytes []byte
			switch src.(type) {
			case []byte:
				srcBytes = src.([]byte)
			case string:
				srcBytes = []byte(src.(string))
			default:
				srcBytes = []byte(fmt.Sprint(src))
			}
			md5Result := md5.Sum(srcBytes)
			return strings.ToUpper(hex.EncodeToString(md5Result[:]))
		},
		"sprintf": func(format string, args ...interface{}) string {
			return fmt.Sprintf(format, args...)
		},
	}
}

func getLogMethod(logf func(string, ...interface{}) bool) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		vm := call.Otto
		args := call.ArgumentList
		if len(args) < 1 {
			panic(vm.MakeTypeError("arguments length must >= 1"))
		}
		logFmt, err := args[0].ToString()
		if err != nil {
			panic(vm.MakeTypeError("arguments error"))
		}
		if len(args) == 1 {
			logf(call.CallerLocation() + "|" + logFmt)
		} else if len(args) > 1 {
			fmtArgs := make([]interface{}, len(args)-1)
			for i := 1; i < len(args); i++ {
				if fmtArgs[i-1], err = args[i].Export(); err != nil {
					panic(vm.MakeTypeError("arguments error"))
				}
			}
			logf(call.CallerLocation()+"|"+logFmt, fmtArgs...)
		}
		return otto.UndefinedValue()
	}
}

func GetLoggerModule() (string, interface{}) {
	return "logger", map[string]interface{}{
		"debug": getLogMethod(niuhe.LogDebug),
		"info":  getLogMethod(niuhe.LogInfo),
		"error": getLogMethod(niuhe.LogError),
	}

}
