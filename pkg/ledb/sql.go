package ledb

import (
	"errors"

	odbclib "github.com/vadv/gopher-lua-libs/db"
	lua "github.com/yuin/gopher-lua"
)

func Open(L *lua.LState, driver, connString string) (db lua.LValue, goerr error) {
	L.CallByParam(lua.P{
		Fn:   L.NewFunction(odbclib.Loader),
		NRet: 1,
	})
	L.Pop(1)

	config := L.NewTable()
	config.RawSetString("shared", lua.LTrue)
	args := []lua.LValue{lua.LString(driver), lua.LString(connString), config}
	L.CallByParam(lua.P{
		Fn:   L.NewFunction(odbclib.Open),
		NRet: 2,
	}, args...)
	db = L.Get(1)
	err := L.Get(2)
	if lua.LVAsBool(err) {
		return nil, errors.New(err.String())
	}
	return
}
