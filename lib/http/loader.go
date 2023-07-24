package http

// Package http implements golang package http utility functionality for lua.

import (
	lua "github.com/yuin/gopher-lua"
)

// Preload adds http to the given Lua state's package.preload table. After it
// has been preloaded, it can be loaded using require:
//
//	local http = require("http")
func Preload(L *lua.LState) {
	L.PreloadModule("http", Loader)
}

// Loader is the module loader function.
func Loader(L *lua.LState) int {
	http_client_ud := L.NewTypeMetatable(`http_client_ud`)
	L.SetField(http_client_ud, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"doreq": DoRequest,
	}))

	http_request_ud := L.NewTypeMetatable(`http_request_ud`)
	L.SetField(http_request_ud, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"auth": Auth,
		"set":  Set,
	}))

	t := L.NewTable()
	L.SetFuncs(t, api)
	L.Push(t)
	return 1
}

var api = map[string]lua.LGFunction{
	"new":    New,
	"newreq": NewRequest,
	"req":    Request,
}
