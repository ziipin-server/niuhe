package db

import (
	"strconv"
	"sync"

	"github.com/go-xorm/xorm"
)

type DB struct {
	engine   *xorm.Engine
	session  *xorm.Session
	lock     sync.Mutex
	initOnce sync.Once
	txLevel  int
}

func NewDB(engine *xorm.Engine) *DB {
	return &DB{
		engine: engine,
	}
}

func (db *DB) GetDB() *xorm.Session {
	db.initOnce.Do(func() {
		db.session = db.engine.NewSession()
	})
	return db.session
}

func (db *DB) Atom(fn func() error) error {
	session := db.GetDB()
	db.lock.Lock()
	if db.txLevel > 0 {
		session.Exec("SAVEPOINT SP_" + strconv.Itoa(db.txLevel))
	} else {
		session.Begin()
	}
	db.txLevel++
	db.lock.Unlock()

	err := fn()

	var dberr error
	db.lock.Lock()
	db.txLevel--
	if db.txLevel > 0 {
		if err != nil {
			_, dberr = session.Exec("ROLLBACK TO SP_" + strconv.Itoa(db.txLevel))
		} else {
			_, dberr = session.Exec("RELEASE SAVEPOINT SP_" + strconv.Itoa(db.txLevel))
		}
	} else {
		if err != nil {
			dberr = session.Rollback()
		} else {
			dberr = session.Commit()
		}
	}
	db.lock.Unlock()
	if dberr != nil {
		panic(dberr)
	}
	return err

}

func (db *DB) Close() {
	if db.session != nil {
		db.lock.Lock()
		if db.session != nil {
			db.session.Close()
		}
		db.lock.Unlock()
	}
}
