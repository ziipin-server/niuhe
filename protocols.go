package niuhe

import "reflect"

// Json API protocol

type jsonApiProtocol struct {
	DefaultApiProtocol
}

func (self *jsonApiProtocol) Read(c *Context, reqValue reflect.Value) error {
	return c.BindJSON(reqValue.Interface())
}

func (self *jsonApiProtocol) Write(c *Context, rsp reflect.Value, err error) error {
	return self.DefaultApiProtocol.Write(c, rsp, err)
}

var JsonApiProtocolFactory = ApiProtocolFactoryFunc(func() IApiProtocol {
	return (*jsonApiProtocol)(nil)
})
