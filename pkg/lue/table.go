package lue

import lua "github.com/yuin/gopher-lua"

func (e *Engine) TGet(t *lua.LTable, k string) lua.LValue {
	return e.L.GetField(t, k)
}

func (e *Engine) TSet(t *lua.LTable, k string, v lua.LValue) {
	e.L.SetField(t, k, v)
}

func (e *Engine) TString(t *lua.LTable, k string) string {
	return string(e.L.GetField(t, k).(lua.LString))
}

func (e *Engine) TInt(t *lua.LTable, k string) int {
	return int(e.L.GetField(t, k).(lua.LNumber))
}

func (e *Engine) TNumber(t *lua.LTable, k string) float64 {
	return float64(e.L.GetField(t, k).(lua.LNumber))
}

func (e *Engine) TBool(t *lua.LTable, k string) bool {
	return lua.LVAsBool(e.L.GetField(t, k))
}

func (e *Engine) TTable(t *lua.LTable, k string) *lua.LTable {
	return e.L.GetField(t, k).(*lua.LTable)
}

func (e *Engine) TFun(t *lua.LTable, k string) *lua.LFunction {
	return e.L.GetField(t, k).(*lua.LFunction)
}

func (e *Engine) TData(t *lua.LTable, k string) any {
	return e.L.GetField(t, k).(*lua.LUserData).Value
}