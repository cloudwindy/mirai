package luaengine

import (
	"mirai/pkg/dir"
	"mirai/pkg/luaextlib"

	lua "github.com/yuin/gopher-lua"
)

// Package info
const (
	LuaVersion = lua.LuaVersion
)

// Package variables
var (
	DefaultIndex = "index.lua"
)

type Engine struct {
	L    *lua.LState
	path string
}

func New(path string, env *lua.LTable) *Engine {
	L := lua.NewState()
	luaextlib.OpenLib(L)
	L.SetGlobal("env", env)
	return &Engine{L: L, path: path}
}

type Factory func(L *lua.LState) lua.LValue

func (e *Engine) Register(name string, factory Factory) {
	e.L.SetGlobal(name, factory(e.L))
}

func (e *Engine) Run() (err error) {
	file, _, err := dir.Index(e.path, DefaultIndex)
	if err != nil {
		return
	}
	if err = e.L.DoFile(file); err != nil {
		return
	}
	return
}
