package niuhe

import (
	"sync"

	"github.com/gorilla/sessions"
)

type _SessCtrl struct {
	Store          sessions.Store
	Session        *sessions.Session
	getSessionOnce sync.Once
	Name           string
	Modified       bool
}

func (sc *_SessCtrl) Save(c *Context) error {
	if !sc.Modified {
		return nil
	}
	if err := sc.Store.Save(c.Request, c.Writer, sc.Session); err != nil {
		return err
	}
	sc.Modified = false
	return nil
}

func (sc *_SessCtrl) Set(c *Context, key string, value interface{}) {
	sc.getSessionOnce.Do(func() {
		var err error
		if sc.Session, err = sc.Store.Get(c.Request, sc.Name); err != nil {
			panic(err)
		}
	})
	sc.Session.Values[key] = value
	sc.Modified = true
}

func (sc *_SessCtrl) Get(c *Context, key string) interface{} {
	sc.getSessionOnce.Do(func() {
		var err error
		if sc.Session, err = sc.Store.Get(c.Request, sc.Name); err != nil {
			panic(err)
		}
	})
	return sc.Session.Values[key]
}

func (sc *_SessCtrl) Del(c *Context, key string) {
	sc.getSessionOnce.Do(func() {
		var err error
		if sc.Session, err = sc.Store.Get(c.Request, sc.Name); err != nil {
			panic(err)
		}
	})
	delete(sc.Session.Values, key)
}

func SessionMiddleware(getStore func() sessions.Store, name string) func(*Context) {
	store := getStore()
	return func(c *Context) {
		c.sessCtrl.Store = store
		c.sessCtrl.Name = name
		c.Next()
		if err := c.sessCtrl.Save(c); err != nil {
			panic(err)
		}
	}
}
