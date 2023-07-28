package art

import (
	lua "github.com/yuin/gopher-lua"
)

func Preload(L *lua.LState) {
	L.PreloadModule("art", Loader)
}

func Loader(L *lua.LState) int {
	L.Push(L.NewFunction(New))
	return 1
}
