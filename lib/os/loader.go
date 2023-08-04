package os

import (
	lua "github.com/yuin/gopher-lua"
)

// Load adds os to the Lua's os table.
func Load(L *lua.LState) {
	modOs := L.RegisterModule(lua.OsLibName, nil).(*lua.LTable)
	L.SetFuncs(modOs, api)
}

var api = map[string]lua.LGFunction{
	"system":   System,
	"read":     Read,
	"write":    Write,
	"stat":     Stat,
	"mkdir":    Mkdir,
	"tmpdir":   TmpDir,
	"hostname": Hostname,
	"pagesize": Pagesize,
}
