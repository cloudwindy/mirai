package ledb

import (
	"os"
	"path"

	"mirai/lib/odbc"
	"mirai/pkg/config"
	"mirai/pkg/lue"

	lua "github.com/yuin/gopher-lua"
)

func New(c config.DB) lue.Module {
	return func(E *lue.Engine) lua.LValue {
		L := E.L
		L.Pop(odbc.Loader(L))
		dbc := odbc.Config{
			Driver:     c.Driver,
			ConnString: c.Conn,
			Shared:     true,
		}
		// open db in protected mode
		pdb, err := odbc.Open(dbc)
		if err != nil {
			L.RaiseError("db open: %v", err)
		}
		mt := L.NewTypeMetatable("db_ud")
		index := L.GetField(mt, "__index").(*lua.LTable)
		index.RawSetString("sqlpath", lua.LString(c.SQLPath))
		L.SetFuncs(index, map[string]lua.LGFunction{
			"loadsql": LoadSQL,
		})

		return E.Class("db_ud", pdb)
	}
}

func LoadSQL(L *lua.LState) int {
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
