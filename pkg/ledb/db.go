package ledb

import (
	"os"
	"path"

	"mirai/pkg/config"
	"mirai/pkg/lue"
	"mirai/pkg/lut"

	lua "github.com/yuin/gopher-lua"
)

func New(c config.DB) lue.Module {
	return func(E *lue.Engine) lua.LValue {
		L := E.L
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
		db.RawSetString("sqlpath", lua.LString(c.SQLPath))
		L.SetFuncs(db, map[string]lua.LGFunction{
			"loadfile": LoadFile,
			"query":    lut.Unprotect(query, pdb, 2),
			"exec":     lut.Unprotect(exec, pdb, 2),
			"command":  lut.Unprotect(command, pdb, 2),
			"stmt":     unprotectStmt(stmt, pdb),
			"close":    lut.Unprotect(close, pdb, 1),
		})

		return db
	}
}

func LoadFile(L *lua.LState) int {
	db := L.CheckTable(1)
	name := L.CheckString(2)
	sqlpath := L.GetField(db, "sqlpath")
	sqlfile := path.Join(lua.LVAsString(sqlpath), name+".sql")
	sql, err := os.ReadFile(sqlfile)
	if err != nil {
		L.RaiseError("db loadfile: %v", err)
	}
	L.Pop(L.GetTop())
	L.CallByParam(lua.P{
		Fn:   L.GetField(db, "exec").(*lua.LFunction),
		NRet: 1,
	}, db, lua.LString(sql))
	return 1
}

func unprotectStmt(fn *lua.LFunction, self lua.LValue) lua.LGFunction {
	stmt := lut.Unprotect(fn, self, 2)
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
			"query": lut.Unprotect(query, pstmt, 2),
			"exec":  lut.Unprotect(exec, pstmt, 2),
			"close": lut.Unprotect(close, pstmt, 1),
		})

		L.Push(upstmt)
		return 1
	}
}
