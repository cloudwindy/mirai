package lecli

import (
	"mirai/pkg/lue"

	lua "github.com/yuin/gopher-lua"
)

type Print func(a ...any)

type Console struct {
	colors map[string]Print
}

func checkCli(L *lua.LState, n int) *Console {
	ud := L.CheckUserData(n)
	cli, ok := ud.Value.(*Console)
	if !ok {
		L.ArgError(n, "expected type Console")
	}
	return cli
}

func New(colors map[string]Print) lue.Module {
	return func(E *lue.Engine) lua.LValue {
		L := E.L
		reg := L.NewTable()

		regFuncs := map[string]lua.LGFunction{}
		for name := range colors {
			regFuncs[name] = colorMethod(name)
		}
		L.SetFuncs(reg, regFuncs)

		return reg
	}
}

func colorMethod(name string) lua.LGFunction {
	return func(L *lua.LState) int {
		cli := checkCli(L, 1)
		str := make([]any, 0, L.GetTop()-1)
		for i := 1; i <= L.GetTop(); i++ {
			str = append(str, L.ToString(i))
		}
		cli.colors[name](str...)
		return 0
	}
}
