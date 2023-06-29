package config

import (
	"mirai/pkg/dir"
	"mirai/pkg/luaextlib"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

var DefaultIndex = "project.lua"

type Config struct {
	Listen  string
	Editing bool
	Data    string
	Root    string
	Index   string
	DB      DB
	Env     *lua.LTable `gluamapper:"-"`
}

type DB struct {
	Driver string
	Conn   string
}

func Parse(configFile string) (c Config, err error) {
	L := lua.NewState()
	defer L.Close()
	luaextlib.OpenLib(L)

	file, isDir, err := dir.Index(configFile, DefaultIndex)
	if err != nil {
		return
	}
	if err = L.DoFile(file); err != nil {
		return
	}
	t := L.CheckTable(1)
	if err = gluamapper.Map(t, &c); err != nil {
		return
	}
	if c.Index == "" && isDir {
		c.Index = configFile
	}
	if env, ok := t.RawGetString("env").(*lua.LTable); ok {
		c.Env = env
	}
	return
}
