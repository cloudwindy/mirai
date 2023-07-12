package lutil

import lua "github.com/yuin/gopher-lua"

func Unprotect(fn *lua.LFunction, self lua.LValue, nret int) lua.LGFunction {
	return func(L *lua.LState) int {
		params := make([]lua.LValue, 0)
		for i := 1; i <= L.GetTop(); i++ {
			params = append(params, L.Get(i))
		}
		params[0] = self
		L.CallByParam(lua.P{
			Fn:   fn,
			NRet: nret,
		}, params...)
		err := L.Get(-1)
		if err != lua.LNil {
			L.RaiseError(lua.LVAsString(err))
		}
		ret := 0
		for i := 1; i <= nret-1; i++ {
			L.Push(L.Get(i))
			ret += 1
		}
		return ret
	}
}
