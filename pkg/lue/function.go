package lue

import lua "github.com/yuin/gopher-lua"

func (e *Engine) CallLFun(lf *lua.LFunction, env *lua.LTable, nret int, params ...lua.LValue) error {
	nf := *lf
	lf = &nf
	e.Clear()
	return e.L.CallByParam(lua.P{
		Fn:      &nf,
		NRet:    nret,
		Protect: true,
	}, params...)
}

func (e *Engine) LFun(fn Fun) *lua.LFunction {
	return e.L.NewFunction(e.LGFun(fn))
}

func (e *Engine) LGFun(fn Fun) lua.LGFunction {
	return func(*lua.LState) int {
		return fn(e)
	}
}

// Get all arguments
func (e *Engine) Arguments() []lua.LValue {
	params := make([]lua.LValue, 0, e.Top())
	for i := 1; i <= e.Top(); i++ {
		params = append(params, e.Get(i))
	}
	return params
}
