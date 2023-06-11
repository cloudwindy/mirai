package pwdchecker

import (
	lua "github.com/yuin/gopher-lua"
)

// Preload adds pwdchecker to the given Lua state's package.preload table. After it
// has been preloaded, it can be loaded using require:
//
//	local pwdchecker = require("pwdchecker")
func Preload(L *lua.LState) {
	L.PreloadModule("pwdchecker", Loader)
}

// Loader is the module loader function.
func Loader(L *lua.LState) int {
	t := L.NewTable()
	L.SetFuncs(t, api)

	mt := L.NewTable()
	mt.RawSetString("__call", L.NewFunction(CallCheck))
	L.SetMetatable(t, mt)

	L.Push(t)
	return 1
}

var api = map[string]lua.LGFunction{
	"check": Check,
}
