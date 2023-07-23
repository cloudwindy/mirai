package http

import (
	"io"

	lua "github.com/yuin/gopher-lua"
)

func newLuaFileReader(L *lua.LState, f lua.LValue) (r io.Reader) {
	read := L.GetField(f, "read").(*lua.LFunction)
	return &luaFileReader{
		f:  f,
		fn: read,
		l:  L,
	}
}

type luaFileReader struct {
	f  lua.LValue
	fn *lua.LFunction
	l  *lua.LState
}

func (r *luaFileReader) Read(p []byte) (n int, err error) {
	r.l.Pop(r.l.GetTop())
	r.l.CallByParam(lua.P{
		Fn:   r.fn,
		NRet: 1,
	}, lua.LNumber(len(p)))
	s := r.l.CheckString(1)
	return copy([]byte(s), p), nil
}
