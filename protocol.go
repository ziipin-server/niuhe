package niuhe

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"reflect"

	"github.com/ziipin-server/zpform"
)

type isCustomRoot interface {
	ThisIsACustomRoot()
}

type IApiProtocol interface {
	Read(*Context, reflect.Value) error
	Write(*Context, reflect.Value, error) error
}

type DefaultApiProtocol struct{}

func (self DefaultApiProtocol) Read(c *Context, reqValue reflect.Value) error {
	req := c.Request
	bindings := zpform.GetReflectedBindings(reqValue)
	if len(bindings) == 0 {
		req.ParseForm()
		return nil
	}
	if req.Method == http.MethodPost && req.Body != nil {
		if mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type")); err == nil && mediaType == "application/json" {
			// Preserve the bytes consumed during binding so the handler can read the full body.
			original := req.Body
			var consumed bytes.Buffer
			body := &struct {
				io.Reader
				io.Closer
			}{io.TeeReader(original, &consumed), original}
			req.Body = body
			defer func() {
				body.Reader = io.MultiReader(&consumed, original)
			}()
		}
	}
	if err := zpform.ReadForm(req, bindings...); err != nil {
		return NewCommError(-1, err.Error())
	}
	return nil
}

func (self DefaultApiProtocol) Write(c *Context, rsp reflect.Value, err error) error {
	rspInst := rsp.Interface()
	if _, ok := rspInst.(isCustomRoot); ok {
		c.JSON(200, rspInst)
	} else {
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
				"data":   rspInst,
			}
		}
		c.JSON(200, response)
	}
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
