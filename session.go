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

func (sc *_SessCtrl) initOnce(c *Context) *_SessCtrl {
	sc.getSessionOnce.Do(func() {
		var err error
		if sc.Session, err = sc.Store.Get(c.Request, sc.Name); err != nil {
			panic(err)
		}
	})
	return sc
}

func (sc *_SessCtrl) Save(c *Context) error {
	if !sc.Modified || sc.Session == nil {
		return nil
	}
	if err := sc.Store.Save(c.Request, c.Writer, sc.Session); err != nil {
		return err
	}
	sc.Modified = false
	return nil
}

func (sc *_SessCtrl) MustSave(c *Context) {
	if err := sc.Save(c); err != nil {
		panic(err)
	}
}

func (sc *_SessCtrl) Set(c *Context, key string, value interface{}) {
	sc.initOnce(c).Session.Values[key] = value
	sc.Modified = true
}

func (sc *_SessCtrl) Get(c *Context, key string) interface{} {
	return sc.initOnce(c).Session.Values[key]
}

func (sc *_SessCtrl) Del(c *Context, key string) {
	delete(sc.initOnce(c).Session.Values, key)
}

func SessionMiddleware(name string, newStoreFn func() sessions.Store) func(*Context) {
	var store sessions.Store
	var initStoreOnce sync.Once
	return func(c *Context) {
		initStoreOnce.Do(func() {
			store = newStoreFn()
		})
		c.sessCtrl.Store = store
		c.sessCtrl.Name = name
		c.Next()
		c.sessCtrl.MustSave(c)
	}
}
