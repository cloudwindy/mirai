package main

import (
	"mirai/lib/lextlib"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

type Config struct {
	Listen    string
	Editing   bool
	ScriptDir string
	RootDir   string
	env       *lua.LTable
}

func GetConfig(luaFile string) Config {
	L := lua.NewState()
	defer L.Close()
	lextlib.OpenLib(L)

	if err := L.DoFile(ConfigFile); err != nil {
		panic(err)
	}
	t := L.CheckTable(1)
	c := new(Config)
	if env, ok := t.RawGetString("env").(*lua.LTable); ok {
		c.env = env
	}
	if err := gluamapper.Map(t, c); err != nil {
		panic(err)
	}
	return *c
}
