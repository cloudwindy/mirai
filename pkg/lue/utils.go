package lue

import (
	lua "github.com/yuin/gopher-lua"
)

func (e *Engine) Error(format string, args ...any) {
	e.L.RaiseError(format, args...)
}

func (e *Engine) MapFuncs(funs map[string]Fun) map[string]lua.LValue {
	dict := make(map[string]lua.LValue)
	for name, fun := range funs {
		dict[name] = e.LFun(fun)
	}
	return dict
}

func (e *Engine) SetFuncs(tb *lua.LTable, funs map[string]Fun) *lua.LTable {
	if tb == nil {
		tb = e.L.NewTable()
	}
	for k, v := range funs {
		tb.RawSetString(k, e.LFun(v))
	}
	return tb
}

func (e *Engine) SetFields(tb *lua.LTable, fields map[string]lua.LValue) *lua.LTable {
	if tb == nil {
		tb = e.L.NewTable()
	}
	for k, v := range fields {
		tb.RawSetString(k, v)
	}
	return tb
}

func (e *Engine) SetDict(tb *lua.LTable, dict map[string]string) *lua.LTable {
	if tb == nil {
		tb = e.L.NewTable()
	}
	for k, v := range dict {
		tb.RawSetString(k, lua.LString(v))
	}
	return tb
}
