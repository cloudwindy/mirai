package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	lua "github.com/yuin/gopher-lua"
)

func Parse(L *lua.LState) int {
	buf := new(bytes.Buffer)
	md := L.CheckString(1)
	if err := goldmark.Convert([]byte(md), buf); err != nil {
		L.RaiseError("%v", err)
	}
	L.Push(lua.LString(buf.String()))
	return 1
}
