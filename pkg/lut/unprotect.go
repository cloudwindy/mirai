package lut

import lua "github.com/yuin/gopher-lua"

func Unprotect(fn *lua.LFunction, self lua.LValue, nret int) lua.LGFunction {
	return func(L *lua.LState) int {
		L.Replace(1, self)
		L.Call(L.GetTop(), nret)
		err := L.Get(-1)
		if err != lua.LNil {
			L.RaiseError(lua.LVAsString(err))
		}
		for i := 1; i <= nret-1; i++ {
			L.Push(L.Get(i))
		}
		return nret - 1
	}
}
