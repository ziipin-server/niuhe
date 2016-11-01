package niuhe

import "reflect"

// Json API protocol

type jsonApiProtocol struct {
	DefaultApiProtocol
}

func (self jsonApiProtocol) Read(c *Context, reqValue reflect.Value) error {
	return c.BindJSON(reqValue.Interface())
}

var JsonApiProtocolFactory = ApiProtocolFactoryFunc(func() IApiProtocol {
	return (*jsonApiProtocol)(nil)
})
