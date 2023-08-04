package io

import (
	lua "github.com/yuin/gopher-lua"
)

// Load adds io to the Lua's io table.
func Load(L *lua.LState) {
	modIo := L.RegisterModule(lua.IoLibName, nil).(*lua.LTable)
	L.SetFuncs(modIo, api)
}

var api = map[string]lua.LGFunction{
	"copy": Copy,
}
