package pwdchecker

import (
	lua "github.com/yuin/gopher-lua"
)

var MinEntropy float64 = 60

func Check(L *lua.LState) int {
	password := L.CheckString(1)
	err := Validate(password, MinEntropy)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LTrue)
	return 1
}
