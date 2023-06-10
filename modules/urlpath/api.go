package urlpath

import (
	"github.com/ucarion/urlpath"
	lua "github.com/yuin/gopher-lua"
)

func New(L *lua.LState) int {
	path := L.CheckString(1)

	p := urlpath.New(path)
	ud := L.NewUserData()
	ud.Value = p

	mt := L.NewTable()
	L.SetMetatable(ud, mt)
	
	exports := L.NewTable()
	L.SetFuncs(exports, urlPathExports)
	mt.RawSetString("__index", exports)

	L.Push(ud)
	return 1
}

var urlPathExports = map[string]lua.LGFunction{
	"match": Match,
}

func Match(L *lua.LState) int {
	path := CheckPath(L, 1)
	s := L.CheckString(2)
	m, ok := path.Match(s)
	if !ok {
		L.Push(lua.LNil)
		return 1
	}
	params := L.NewTable()
	for _, v := range m.Params {
		params.Append(lua.LString(v))
	}
	trailing := lua.LString(m.Trailing)
	L.Push(params)
	L.Push(trailing)
	return 2
}

func CheckPath(L *lua.LState, n int) *urlpath.Path {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(*urlpath.Path); ok {
		return v
	}
	L.ArgError(1, "urlpath expected")
	return nil
}
