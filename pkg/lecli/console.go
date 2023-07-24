package lecli

import (
	"github.com/cloudwindy/mirai/pkg/lue"
	"github.com/inancgumus/screen"
	lua "github.com/yuin/gopher-lua"
)

type Print func(format string, a ...any)

func New(colors map[string]Print) lue.Module {
	return func(E *lue.Engine) lua.LValue {
		L := E.L
		cli := L.NewTable()

		regFuncs := map[string]lua.LGFunction{}
		for name, p := range colors {
			regFuncs[name] = colorMethod(p)
		}
		L.SetFuncs(cli, regFuncs)

		E.SetFuncs(cli, map[string]lue.Fun{
			"clear": clear,
		})

		return cli
	}
}

func colorMethod(p Print) lua.LGFunction {
	return func(L *lua.LState) int {
		format := L.CheckString(1)
		str := make([]any, 0, L.GetTop()-1)
		for i := 2; i <= L.GetTop(); i++ {
			str = append(str, L.ToStringMeta(L.Get(i)).String())
		}
		p(format, str...)
		return 0
	}
}

func clear(E *lue.Engine) int {
	screen.Clear()
	screen.MoveTopLeft()
	return 0
}
