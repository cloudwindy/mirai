package art

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zs5460/art"
)

func New(L *lua.LState) int {
	text := L.CheckString(1)
	text = art.String(text)
	L.Push(lua.LString(text))
	return 1
}
