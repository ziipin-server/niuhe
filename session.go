package niuhe

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

const (
	__SESSION_OPTIONS_KEY__ = "__session_options_key__"
)

type _SessCtrl struct {
	Store          sessions.Store
	Session        *sessions.Session
	getSessionOnce sync.Once
	Name           string
	Modified       bool
	OptionLoaded   bool
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
	sc.Modified = true
}

func (sc *_SessCtrl) GetOptions(c *Context) *sessions.Options {
	sc.initOnce(c)
	if !sc.OptionLoaded {
		bOpts, _ := sc.Session.Values[__SESSION_OPTIONS_KEY__].([]byte)
		opts, ok := new(sessions.Options), false
		if len(bOpts) > 0 {
			if err := json.Unmarshal(bOpts, opts); err != nil {
				fmt.Println(err.Error())
			} else {
				ok = true
			}
		}
		if ok {
			sc.Session.Options = opts
		}
		sc.OptionLoaded = true
	}
	return sc.Session.Options
}

func (sc *_SessCtrl) SetOptions(c *Context, opts *sessions.Options) {
	opts = &sessions.Options{
		Path:     opts.Path,
		Domain:   opts.Domain,
		MaxAge:   opts.MaxAge,
		Secure:   opts.Secure,
		HttpOnly: opts.HttpOnly,
	}
	sc.initOnce(c).Session.Options = opts
	bOpts, _ := json.Marshal(opts)
	sc.Session.Values[__SESSION_OPTIONS_KEY__] = bOpts
	sc.Modified = true
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

func GetCodecs(c *Context, name string, keyPairs string) string {
	codecs := securecookie.CodecsFromPairs([]byte(keyPairs))
	sess, err := c.Context.Request.Cookie(name)
	if err != nil {
		fmt.Println("Get sess err: " + err.Error())
		return ""
	}
	s := new(string)
	securecookie.DecodeMulti(name, sess.Value, s, codecs...)
	return *s
}
