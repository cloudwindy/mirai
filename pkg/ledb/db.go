package ledb

import (
	"mirai/pkg/config"
	"mirai/pkg/luaengine"

	lua "github.com/yuin/gopher-lua"
)

func New(c config.DB) luaengine.Factory {
	return func(L *lua.LState) lua.LValue {
		// open db in protected mode
		pdb, err := Open(L, c.Driver, c.Conn)
		if err != nil {
			L.RaiseError("db open failed: %v", err)
		}

		query := L.GetField(pdb, "query").(*lua.LFunction)
		exec := L.GetField(pdb, "exec").(*lua.LFunction)
		stmt := L.GetField(pdb, "stmt").(*lua.LFunction)
		close := L.GetField(pdb, "close").(*lua.LFunction)
		db := L.NewTable()
		L.SetFuncs(db, map[string]lua.LGFunction{
			"query": unprotectRet1(query),
			"exec":  unprotectRet1(exec),
			"stmt":  unprotectStmt(stmt),
			"close": unprotectRet0(close),
		})

		return db
	}
}

func unprotectRet0(fn *lua.LFunction) lua.LGFunction {
	return func(L *lua.LState) int {
		params := make([]lua.LValue, 0)
		for i := 1; i <= L.GetTop(); i++ {
			params = append(params, L.Get(i))
		}
		L.CallByParam(lua.P{
			Fn:   fn,
			NRet: 1,
		}, params...)
		err := L.Get(1)
		if err != lua.LNil {
			L.RaiseError(lua.LVAsString(err))
		}
		return 0
	}
}

func unprotectRet1(fn *lua.LFunction) lua.LGFunction {
	return func(L *lua.LState) int {
		params := make([]lua.LValue, 0)
		for i := 1; i <= L.GetTop(); i++ {
			params = append(params, L.Get(i))
		}
		L.CallByParam(lua.P{
			Fn:   fn,
			NRet: 2,
		}, params...)
		ret := L.Get(1)
		err := L.Get(2)
		if err != lua.LNil {
			L.RaiseError(lua.LVAsString(err))
		}
		L.Push(ret)
		return 1
	}
}

func unprotectStmt(fn *lua.LFunction) lua.LGFunction {
	stmt := unprotectRet1(fn)
	return func(L *lua.LState) int {
		stmt(L)
		// prepared statement
		pstmt := L.Get(1)

		// unprotected prepared statement
		upstmt := L.NewTable()

		query := L.GetField(pstmt, "query").(*lua.LFunction)
		exec := L.GetField(pstmt, "exec").(*lua.LFunction)
		close := L.GetField(pstmt, "close").(*lua.LFunction)
		L.SetFuncs(upstmt, map[string]lua.LGFunction{
			"query": unprotectRet1(query),
			"exec":  unprotectRet1(exec),
			"close": unprotectRet0(close),
		})

		L.Push(upstmt)
		return 1
	}
}
