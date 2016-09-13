package niuhe

import (
	"reflect"

	"github.com/ziipin-server/zpform"
)

type IApiProtocol interface {
	Read(*Context, reflect.Value) error
	Write(*Context, reflect.Value, error) error
}

type DefaultApiProtocol struct{}

func (self DefaultApiProtocol) Read(c *Context, reqValue reflect.Value) error {
	if err := zpform.ReadReflectedStructForm(c.Request, reqValue); err != nil {
		return NewCommError(-1, err.Error())
	}
	return nil
}

func (self DefaultApiProtocol) Write(c *Context, rsp reflect.Value, err error) error {
	var response map[string]interface{}
	if err != nil {
		if commErr, ok := err.(ICommError); ok {
			response = map[string]interface{}{
				"result":  commErr.GetCode(),
				"message": commErr.GetMessage(),
			}
			if commErr.GetCode() == 0 {
				response["data"] = rsp.Interface()
			}
		} else {
			response = map[string]interface{}{
				"result":  -1,
				"message": err.Error(),
			}
		}
	} else {
		response = map[string]interface{}{
			"result": 0,
			"data":   rsp.Interface(),
		}
	}
	c.JSON(200, response)
	return nil
}

type IApiProtocolFactory interface {
	GetProtocol() IApiProtocol
}

type ApiProtocolFactoryFunc func() IApiProtocol

func (f ApiProtocolFactoryFunc) GetProtocol() IApiProtocol {
	return f()
}

var defaultApiProtocol *DefaultApiProtocol

func GetDefaultApiProtocol() IApiProtocol {
	LogDebug("GetDefaultApiProtocol")
	return defaultApiProtocol
}

var defaultApiProtocolFactory IApiProtocolFactory

func GetDefaultProtocolFactory() IApiProtocolFactory {
	return defaultApiProtocolFactory
}

func SetDefaultProtocolFactory(pf IApiProtocolFactory) {
	defaultApiProtocolFactory = pf
}

func init() {
	defaultApiProtocol = &DefaultApiProtocol{}
	defaultApiProtocolFactory = ApiProtocolFactoryFunc(GetDefaultApiProtocol)
}
