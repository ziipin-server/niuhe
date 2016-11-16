package db

import (
	"math/rand"
	"strconv"
	"sync"

	"github.com/go-xorm/xorm"
)

type DB struct {
	engine       *xorm.Engine
	slaveEngines []*xorm.Engine
	session      *xorm.Session
	slaveSession *xorm.Session
	lock         sync.Mutex
	masterOnce   sync.Once
	slaveOnce    sync.Once
	txLevel      int
}

func NewDB(engine *xorm.Engine) *DB {
	return &DB{
		engine: engine,
	}
}

func NewDBWithSlaves(masterEngine *xorm.Engine, slaveEngines []*xorm.Engine) *DB {
	return &DB{
		engine:       masterEngine,
		slaveEngines: slaveEngines,
	}
}

func (db *DB) GetMasterDB() *xorm.Session {
	db.masterOnce.Do(func() {
		db.session = db.engine.NewSession()
	})
	return db.session
}

func (db *DB) GetSlaveDB() *xorm.Session {
	db.slaveOnce.Do(func() {
		idx := rand.Intn(len(db.slaveEngines)) // if no salve engines, it will panic
		db.slaveSession = db.slaveEngines[idx].NewSession()
	})
	return db.slaveSession
}

func (db *DB) GetDB() *xorm.Session {
	if db.txLevel > 0 || len(db.slaveEngines) < 1 {
		return db.GetMasterDB()
	} else {
		return db.GetSlaveDB()
	}
}

func (db *DB) Atom(fn func() error) error {
	session := db.GetMasterDB()
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
	db.Close() // close session
	return err

}

func (db *DB) Close() {
	if db.session != nil {
		db.lock.Lock()
		if db.session != nil {
			db.session.Close()
		}
		db.session = nil
		db.lock.Unlock()
	}
	if db.slaveSession != nil {
		db.lock.Lock()
		if db.slaveSession != nil {
			db.slaveSession.Close()
		}
		db.lock.Unlock()
	}
}
