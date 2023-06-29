package ledb

import (
	"errors"

	sql "github.com/vadv/gopher-lua-libs/db"
	lua "github.com/yuin/gopher-lua"
)

func Open(L *lua.LState, driver, connString string) (db lua.LValue, err error) {
	config := L.NewTable()
	config.RawSetString("shared", lua.LTrue)
	args := []lua.LValue{lua.LString(driver), lua.LString(connString), config}
	err = L.CallByParam(lua.P{
		Fn:   L.NewFunction(sql.Open),
		NRet: 2,
	}, args...)
	if err != nil {
		return
	}
	db = L.Get(1)
	luaerr := L.Get(2)
	if lua.LVAsBool(luaerr) {
		return nil, errors.New(luaerr.String())
	}
	return
}
