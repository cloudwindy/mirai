package ledb

import (
	"mirai/pkg/config"
	"mirai/pkg/luaengine"

	lua "github.com/yuin/gopher-lua"
)

func New(c config.DB) luaengine.Factory {
	return func(L *lua.LState) lua.LValue {
		db, err := Open(L, c.Driver, c.Conn)
		if err != nil {
			L.RaiseError("db open failed: %v", err)
		}
		return db
	}
}
