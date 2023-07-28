package lue

import (
	lua "github.com/yuin/gopher-lua"
)

func (e *Engine) NewTable() *lua.LTable {
	return e.L.NewTable()
}

func (e *Engine) NewData(val interface{}) lua.LValue {
	ud := e.L.NewUserData()
	ud.Value = val
	return ud
}

// read-write object
func (e *Engine) ReadWrite(getter lua.LGFunction, setter lua.LGFunction) lua.LValue {
	mt := map[string]lua.LGFunction{
		"__index":    getter,
		"__newindex": setter,
	}
	return e.ProxyFuncs(mt)
}

// read only object
func (e *Engine) ReadOnly(getter lua.LGFunction) lua.LValue {
	mt := map[string]lua.LGFunction{
		"__index": getter,
	}
	return e.ProxyFuncs(mt)
}

func (e *Engine) Anonymous(value interface{}, index lua.LValue, newIndex ...lua.LValue) lua.LValue {
	obj := e.NewData(value)
	mt := e.L.NewTable()
	mt.RawSetString("__index", index)
	if len(newIndex) > 0 {
		mt.RawSetString("__newindex", newIndex[0])
	}
	e.L.SetMetatable(obj, mt)
	return obj
}

func (e *Engine) Class(name string, value interface{}, index ...lua.LValue) lua.LValue {
	obj := e.NewData(value)
	mt := e.L.NewTypeMetatable(name)
	if len(index) > 0 {
		mt.RawSetString("__index", index[0])
	}
	e.L.SetMetatable(obj, mt)
	return obj
}

// proxy object with metatable functions
func (e *Engine) ProxyFuncs(mtFuncs map[string]lua.LGFunction) lua.LValue {
	L := e.L
	obj := L.NewUserData()
	mt := L.NewTable()
	L.SetFuncs(mt, mtFuncs)
	L.SetMetatable(obj, mt)
	return obj
}
