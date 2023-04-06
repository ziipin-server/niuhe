package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"

	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/ziipin-server/niuhe"
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

func (db *DB) Atom(fn func() error, ctx ...context.Context) error {
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
		if errors.Is(dberr, context.Canceled) {
			return dberr
		} else {
			// 假如由于context被cancel，由于重复rollback/commit导致的sql.ErrTxDone，不panic
			if len(ctx) > 0 && errors.Is(ctx[0].Err(), context.Canceled) && errors.Is(dberr, sql.ErrTxDone) {
				return dberr
			}
			panic(dberr)
		}
	}
	db.txLevel++
	db.lock.Unlock()

	var err error
	var fnLastSql, txLastSql string
	var fnLastSqlParams, txLastSqlParams []interface{}
	hasPanic := true
	defer func() {
		db.lock.Lock()
		defer db.lock.Unlock()
		db.txLevel--
		if db.txLevel > 0 {
			if err != nil || hasPanic || db.txFailed {
				txLastSql, txLastSqlParams = session.LastSQL()
				_, dberr = session.Exec("ROLLBACK TO SP_" + strconv.Itoa(db.txLevel))
			} else {
				txLastSql, txLastSqlParams = session.LastSQL()
				_, dberr = session.Exec("RELEASE SAVEPOINT SP_" + strconv.Itoa(db.txLevel))
			}
		} else {
			if err != nil || hasPanic || db.txFailed {
				txLastSql, txLastSqlParams = session.LastSQL()
				dberr = session.Rollback()
			} else {
				txLastSql, txLastSqlParams = session.LastSQL()
				dberr = session.Commit()
			}
			session.Close()
			db.session = nil
			db.masterOnce = sync.Once{}
			db.txFailed = false
		}
		if dberr != nil && !errors.Is(err, context.Canceled) {
			lastSql, lastSqlParams := session.LastSQL()
			buf, _ := json.Marshal(map[string]interface{}{
				"db.txLevel":      db.txLevel,
				"err":             err,
				"err_str":         fmt.Sprintf("%v", err),
				"hasPanic":        hasPanic,
				"db.txFaild":      db.txFailed,
				"dberr":           dberr,
				"dberr_str":       fmt.Sprintf("%v", dberr),
				"fnLastSql":       fnLastSql,
				"fnLastSqlParams": fnLastSqlParams,
				"txLastSql":       txLastSql,
				"txLastSqlParams": txLastSqlParams,
				"lastSql":         lastSql,
				"lastSqlParams":   lastSqlParams,
			})
			niuhe.LogWarn("[TxFail] %s\n", string(buf))
			if len(ctx) > 0 && errors.Is(ctx[0].Err(), context.Canceled) && errors.Is(dberr, sql.ErrTxDone) {
				niuhe.LogInfo("[TxWatch] transaction has already been committed or rolled back when context is canceled, err=%v, dberr=%v", err, dberr)
			} else {
				panic(dberr)
			}
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
		fnLastSql, fnLastSqlParams = session.LastSQL()
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
