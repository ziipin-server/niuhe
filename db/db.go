package db

import (
	"math/rand"
	"strconv"
	"sync"

	"github.com/go-sql-driver/mysql"
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
	txFailed     bool
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
		db.txFailed = false
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
	var dberr error
	session := db.GetMasterDB()
	db.lock.Lock()
	if db.txLevel > 0 {
		_, dberr = session.Exec("SAVEPOINT SP_" + strconv.Itoa(db.txLevel))
	} else {
		dberr = session.Begin()
	}
	if dberr != nil {
		db.lock.Unlock()
		panic(dberr)
	}
	db.txLevel++
	db.lock.Unlock()

	var err error
	hasPanic := true
	defer func() {
		db.lock.Lock()
		defer db.lock.Unlock()
		db.txLevel--
		if db.txLevel > 0 {
			if err != nil || hasPanic || db.txFailed {
				_, dberr = session.Exec("ROLLBACK TO SP_" + strconv.Itoa(db.txLevel))
			} else {
				_, dberr = session.Exec("RELEASE SAVEPOINT SP_" + strconv.Itoa(db.txLevel))
			}
		} else {
			if err != nil || hasPanic || db.txFailed {
				dberr = session.Rollback()
			} else {
				dberr = session.Commit()
			}
			session.Close()
			db.session = nil
			db.masterOnce = sync.Once{}
			db.txFailed = false
		}
		if dberr != nil {
			panic(dberr)
		}
	}()
	err = fn()
	if err != nil {
		switch ferr := err.(type) {
		case *mysql.MySQLError:
			switch ferr.Number {
			case 1205: // [MYSQLError] 1205 - Error 1205: Lock wait timeout exceeded; try restarting transaction
				fallthrough
			case 1213: // [MYSQLError] 1213 - Error 1213: Deadlock found when trying to get lock; try restarting transaction
				db.txFailed = true
			}
		}
	}
	hasPanic = false
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
