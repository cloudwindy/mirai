package lutil

import lua "github.com/yuin/gopher-lua"

func Unprotect(fn *lua.LFunction, self lua.LValue, nret int) lua.LGFunction {
	p := lua.P{
		Fn:   fn,
		NRet: nret,
	}
	return func(L *lua.LState) int {
		params := make([]lua.LValue, 0)
		for i := 1; i <= L.GetTop(); i++ {
			params = append(params, L.Get(i))
		}
		params[0] = self
		L.CallByParam(p, params...)
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
