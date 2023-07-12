package ledb

import (
	"mirai/pkg/config"
	"mirai/pkg/luaengine"
	"mirai/pkg/lutil"

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
		command := L.GetField(pdb, "command").(*lua.LFunction)
		close := L.GetField(pdb, "close").(*lua.LFunction)
		db := L.NewTable()
		L.SetFuncs(db, map[string]lua.LGFunction{
			"query":   lutil.Unprotect(query, pdb, 2),
			"exec":    lutil.Unprotect(exec, pdb, 2),
			"command": lutil.Unprotect(command, pdb, 2),
			"stmt":    unprotectStmt(stmt, pdb),
			"close":   lutil.Unprotect(close, pdb, 1),
		})

		return db
	}
}

func unprotectStmt(fn *lua.LFunction, self lua.LValue) lua.LGFunction {
	stmt := lutil.Unprotect(fn, self, 2)
	return func(L *lua.LState) int {
		stmt(L)
		// prepared statement
		pstmt := L.Get(1)
		query := L.GetField(pstmt, "query").(*lua.LFunction)
		exec := L.GetField(pstmt, "exec").(*lua.LFunction)
		close := L.GetField(pstmt, "close").(*lua.LFunction)

		// unprotected prepared statement
		upstmt := L.NewTable()
		L.SetFuncs(upstmt, map[string]lua.LGFunction{
			"query": lutil.Unprotect(query, pstmt, 2),
			"exec":  lutil.Unprotect(exec, pstmt, 2),
			"close": lutil.Unprotect(close, pstmt, 1),
		})

		L.Push(upstmt)
		return 1
	}
}
