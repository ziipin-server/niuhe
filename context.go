package niuhe

import "github.com/gin-gonic/gin"

type Context struct {
	*gin.Context
	index    int8
	handlers []HandlerFunc
	sessCtrl _SessCtrl
}

func newContext(c *gin.Context, middlewares []HandlerFunc) *Context {
	return &Context{Context: c, index: -1, handlers: middlewares}
}

func (c *Context) Next() {
	c.index++
	s := int8(len(c.handlers))
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Abort() {
	c.index = abortIndex
}

// Session segment

func (c *Context) SetSession(key string, value interface{}) {
	c.sessCtrl.Set(c, key, value)
}

func (c *Context) GetSession(key string) interface{} {
	return c.sessCtrl.Get(c, key)
}

func (c *Context) DelSession(key string) {
	c.sessCtrl.Del(c, key)
}

func (c *Context) beforeOutput() {
	c.sessCtrl.MustSave(c)
}

func (c *Context) HTML(code int, name string, obj interface{}) {
	c.beforeOutput()
	c.Context.HTML(code, name, obj)
}

func (c *Context) JSON(code int, obj interface{}) {
	c.beforeOutput()
	c.Context.JSON(code, obj)
}

func (c *Context) IndentedJSON(code int, obj interface{}) {
	c.beforeOutput()
	c.Context.IndentedJSON(code, obj)
}

func (c *Context) XML(code int, obj interface{}) {
	c.beforeOutput()
	c.Context.XML(code, obj)
}

// YAML serializes the given struct as YAML into the response body.
func (c *Context) YAML(code int, obj interface{}) {
	c.beforeOutput()
	c.Context.YAML(code, obj)
}

// String writes the given string into the response body.
func (c *Context) String(code int, format string, values ...interface{}) {
	c.beforeOutput()
	c.Context.String(code, format, values...)
}

// Redirect returns a HTTP redirect to the specific location.
func (c *Context) Redirect(code int, location string) {
	c.beforeOutput()
	c.Context.Redirect(code, location)
}

// Data writes some data into the body stream and updates the HTTP code.
func (c *Context) Data(code int, contentType string, data []byte) {
	c.beforeOutput()
	c.Context.Data(code, contentType, data)
}

// File writes the specified file into the body stream in a efficient way.
func (c *Context) File(filepath string) {
	c.beforeOutput()
	c.Context.File(filepath)
}
