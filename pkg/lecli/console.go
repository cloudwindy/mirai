package lecli

import (
	"mirai/pkg/luaengine"

	lua "github.com/yuin/gopher-lua"
)

func New() luaengine.Factory {
	return func(L *lua.LState) lua.LValue {
		reg := L.NewTable()

		regFuncs := map[string]lua.LGFunction{
			"command": func(L *lua.LState) int {
				return 0
			},
		}
		L.SetFuncs(reg, regFuncs)

		return reg
	}
}
