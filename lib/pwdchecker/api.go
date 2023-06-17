package pwdchecker

import (
	lua "github.com/yuin/gopher-lua"
)

var DefaultMinEntropy float64 = 60

func CallCheck(L *lua.LState) int {
	L.Remove(1)
	return Check(L)
}

func Check(L *lua.LState) int {
	password := L.CheckString(1)
	minEntropy := L.OptNumber(2, lua.LNumber(DefaultMinEntropy))
	err := Validate(password, float64(minEntropy))
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LTrue)
	return 1
}
