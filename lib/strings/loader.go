package strings

import (
	lua "github.com/yuin/gopher-lua"
)

// Load adds string to the Lua's string table.
func Load(L *lua.LState) {
	readerMt := registerStringsReader(L)
	builderMt := registerStringsBuilder(L)

	modString := L.RegisterModule(lua.StringLibName, nil).(*lua.LTable)
	modString.RawSetString("Reader", readerMt)
	modString.RawSetString("Builder", builderMt)
	L.SetFuncs(modString, api)
}

var api = map[string]lua.LGFunction{
	"split":      Split,
	"fields":     Fields,
	"includes":   Contains,
	"startswith": HasPrefix,
	"endswith":   HasSuffix,
	"trim":       Trim,
	"trimspace":  TrimSpace,
	"trimstart":  TrimPrefix,
	"trimend":    TrimSuffix,
}
