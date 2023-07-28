package lue

import lua "github.com/yuin/gopher-lua"

func (e *Engine) Top() int {
	return e.L.GetTop()
}

func (e *Engine) Clear() {
	e.L.Pop(e.L.GetTop())
}

func (e *Engine) Get(n int) lua.LValue {
	return e.L.Get(n)
}

func (e *Engine) String(n int) string {
	return e.L.CheckString(n)
}

func (e *Engine) Int(n int) int {
	return e.L.CheckInt(n)
}

func (e *Engine) Number(n int) float64 {
	return float64(e.L.CheckNumber(n))
}

func (e *Engine) Bool(n int) bool {
	return e.L.ToBool(n)
}

func (e *Engine) Table(n int) *lua.LTable {
	return e.L.CheckTable(n)
}

func (e *Engine) Fun(n int) *lua.LFunction {
	return e.L.CheckFunction(n)
}

func (e *Engine) Data(n int) any {
	ud := e.L.CheckUserData(n)
	return ud.Value
}

func (e *Engine) IsNil(n int) bool {
	return e.L.Get(n) == lua.LNil
}

func (e *Engine) Push(v lua.LValue) {
	e.L.Push(v)
}

func (e *Engine) PushNil() {
	e.L.Push(lua.LNil)
}

func (e *Engine) PushBool(b bool) {
	e.L.Push(lua.LBool(b))
}

func (e *Engine) PushInt(n int) {
	e.L.Push(lua.LNumber(n))
}

func (e *Engine) PushNumber(n float64) {
	e.L.Push(lua.LNumber(n))
}

func (e *Engine) PushString(s string) {
	e.L.Push(lua.LString(s))
}

func (e *Engine) PushData(v any) {
	e.L.Push(e.NewData(v))
}
