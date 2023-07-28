package lecli

import (
	"github.com/cloudwindy/mirai/pkg/lue"
	"github.com/inancgumus/screen"
	lua "github.com/yuin/gopher-lua"
)

type Printf func(format string, a ...any)

func New(colors map[string]Printf) lue.Module {
	return func(E *lue.Engine) lua.LValue {
		cli := E.NewTable()

		regFuncs := map[string]lua.LGFunction{}
		for name, p := range colors {
			regFuncs[name] = colorMethod(p)
		}
		E.L.SetFuncs(cli, regFuncs)
		E.SetFuncs(cli, cliExports)

		return cli
	}
}

func colorMethod(p Printf) lua.LGFunction {
	return func(L *lua.LState) int {
		format := L.CheckString(1)
		str := make([]any, 0, L.GetTop()-1)
		for i := 2; i <= L.GetTop(); i++ {
			str = append(str, L.Get(i))
		}
		p(format, str...)
		return 0
	}
}

var cliExports = map[string]lue.Fun{
	"clear": clear,
}

func clear(E *lue.Engine) int {
	screen.Clear()
	screen.MoveTopLeft()
	return 0
}
