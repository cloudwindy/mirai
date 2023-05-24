package msgpack

import (
	lua "github.com/yuin/gopher-lua"
)

// Preload adds msgpack to the given Lua state's package.preload table. After it
// has been preloaded, it can be loaded using require:
//
//  local msgpack = require("msgpack")
func Preload(L *lua.LState) {
	L.PreloadModule("msgpack", Loader)
}

// Loader is the module loader function.
func Loader(L *lua.LState) int {
	t := L.NewTable()
	L.SetFuncs(t, api)
	L.Push(t)
	return 1
}

var api = map[string]lua.LGFunction{
	"tableIsObject": TableIsObject,
	"decode":        Decode,
	"encode":        Encode,
}
