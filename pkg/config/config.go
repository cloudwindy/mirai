package config

import (
	"mirai/pkg/luaextlib"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

type Root struct {
	Listen    string
	Editing   bool
	DataDir   string
	ScriptDir string
	RootDir   string
	DB        DB
	Env       *lua.LTable `gluamapper:"-"`
}

type DB struct {
	Driver string
	Conn   string
}

func Get(luaFile string) Root {
	L := lua.NewState()
	defer L.Close()
	luaextlib.OpenLib(L)

	if err := L.DoFile(luaFile); err != nil {
		panic(err)
	}
	t := L.CheckTable(1)
	c := new(Root)
	if env, ok := t.RawGetString("env").(*lua.LTable); ok {
		c.Env = env
	}
	if err := gluamapper.Map(t, c); err != nil {
		panic(err)
	}
	return *c
}
