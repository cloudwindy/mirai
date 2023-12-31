package markdown

import lua "github.com/yuin/gopher-lua"

func Preload(L *lua.LState){
	L.PreloadModule("markdown", Loader)
}

func Loader(L *lua.LState) int {
	t := L.NewTable()
	L.SetFuncs(t, api)
	L.Push(t)
	return 1
}

var api = map[string]lua.LGFunction{
	"parse": Parse,
}