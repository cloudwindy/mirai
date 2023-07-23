package re

import (
	lua "github.com/yuin/gopher-lua"
)

// Preload adds re to the given Lua state's package.preload table. After it
// has been preloaded, it can be loaded using require:
//
//	local re = require("re")
func Preload(L *lua.LState) {
	L.PreloadModule("re", Loader)
}

// Loader is the module loader function.
func Loader(L *lua.LState) int {
	regexp_ud := L.NewTypeMetatable(`regexp_ud`)
	L.SetGlobal(`regexp_ud`, regexp_ud)
	L.SetField(regexp_ud, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"match":    CompiledMatch,
		"findall": CompiledFindAll,
	}))

	t := L.NewTable()
	L.SetFuncs(t, api)
	L.Push(t)
	return 1
}

var api = map[string]lua.LGFunction{
	"compile": Compile,
	"match":   Match,
	"findall": FindAll,
}
