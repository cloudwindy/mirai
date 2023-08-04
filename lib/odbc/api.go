// Package db implements golang package db functionality for lua.
package odbc

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

const (
	// max open connections
	MaxOpenConns = 1
)

type luaDB interface {
	constructor(*dbConfig) (luaDB, error)
	getDB() *sql.DB
	closeDB() error
	getTXOptions() *sql.TxOptions
}

type Config struct {
	Driver         string
	ConnString     string
	Shared         bool
	MaxConnections int
	ReadOnly       bool
}

type dbConfig struct {
	connString   string
	sharedMode   bool
	maxOpenConns int
	readOnly     bool
}

var (
	knownDrivers     = make(map[string]luaDB, 0)
	knownDriversLock = &sync.Mutex{}
)

// RegisterDriver register sql driver
func RegisterDriver(driver string, i luaDB) {
	knownDriversLock.Lock()
	defer knownDriversLock.Unlock()

	knownDrivers[driver] = i
}

func checkDB(L *lua.LState, n int) luaDB {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(luaDB); ok {
		return v
	}
	L.ArgError(n, "database expected")
	return nil
}

// Open lua db.open(driver, connection_string, config) returns db_ud
// config table:
//
//	{
//	  shared=false,
//	  max_connections=X,
//	  read_only=false
//	}
func LuaOpen(L *lua.LState) int {
	driver := L.CheckString(1)
	connString := L.CheckString(2)
	var c Config
	if L.GetTop() > 2 {
		config := L.CheckTable(3)
		if err := gluamapper.Map(config, c); err != nil {
			L.RaiseError("%v", err)
		}
	}
	c.Driver = driver
	c.ConnString = connString
	db, err := Open(c)
	if err != nil {
		L.RaiseError("%v", err)
	}
	ud := L.NewUserData()
	ud.Value = db
	L.SetMetatable(ud, L.GetTypeMetatable(`db_ud`))
	L.Push(ud)
	return 1
}

func Open(c Config) (luaDB, error) {
	knownDriversLock.Lock()
	defer knownDriversLock.Unlock()

	driver := c.Driver
	db, ok := knownDrivers[driver]
	if !ok {
		return nil, fmt.Errorf("unknown driver: %s", driver)
	}

	// parse config
	config := &dbConfig{
		connString:   c.ConnString,
		sharedMode:   c.Shared,
		maxOpenConns: c.MaxConnections,
		readOnly:     c.ReadOnly,
	}

	dbIface, err := db.constructor(config)
	if err != nil {
		return nil, err
	}

	return dbIface, nil
}

// Query lua db_ud:query(query) returns {rows = {}, columns = {}}
func Query(L *lua.LState) int {
	dbInterface := checkDB(L, 1)
	query := L.CheckString(2)
	sqlDB := dbInterface.getDB()
	opts := dbInterface.getTXOptions()
	tx, err := sqlDB.BeginTx(context.Background(), opts)
	if err != nil {
		L.RaiseError("%v", err)
	}
	defer tx.Rollback()
	sqlRows, err := tx.Query(query)
	if err != nil {
		L.RaiseError("%v", err)
	}
	defer sqlRows.Close()
	rows, columns, err := parseRows(sqlRows, L)
	if err != nil {
		L.RaiseError("%v", err)
	}
	tx.Commit()
	result := L.NewTable()
	result.RawSetString(`rows`, rows)
	result.RawSetString(`columns`, columns)
	L.Push(result)
	return 1
}

// Exec lua db_ud:exec(query) returns {rows_affected=number, last_insert_id=number}
func Exec(L *lua.LState) int {
	dbInterface := checkDB(L, 1)
	query := L.CheckString(2)
	sqlDB := dbInterface.getDB()
	opts := dbInterface.getTXOptions()
	tx, err := sqlDB.BeginTx(context.Background(), opts)
	if err != nil {
		L.RaiseError("%v", err)
	}
	defer tx.Rollback()
	sqlResult, err := tx.Exec(query)
	if err != nil {
		L.RaiseError("%v", err)
	}
	tx.Commit()
	result := L.NewTable()
	if id, err := sqlResult.LastInsertId(); err == nil {
		result.RawSetString(`last_insert_id`, lua.LNumber(id))
	}
	if aff, err := sqlResult.RowsAffected(); err == nil {
		result.RawSetString(`rows_affected`, lua.LNumber(aff))
	}
	L.Push(result)
	return 1
}

// Command lua db_ud:command(query) returns {rows = {}, columns = {}}
func Command(L *lua.LState) int {
	dbInterface := checkDB(L, 1)
	query := L.CheckString(2)
	sqlDB := dbInterface.getDB()
	sqlRows, err := sqlDB.Query(query)
	if err != nil {
		L.RaiseError("%v", err)
	}
	defer sqlRows.Close()
	rows, columns, err := parseRows(sqlRows, L)
	if err != nil {
		L.RaiseError("%v", err)
	}
	result := L.NewTable()
	result.RawSetString(`rows`, rows)
	result.RawSetString(`columns`, columns)
	L.Push(result)
	return 1
}

// Close lua db_ud:close()
func Close(L *lua.LState) int {
	dbIface := checkDB(L, 1)
	if err := dbIface.closeDB(); err != nil {
		L.RaiseError("%v", err)
	}
	return 0
}
