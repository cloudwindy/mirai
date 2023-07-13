package config

import (
	"io"
	"os"

	"mirai/pkg/lelib"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

var ProjectFileName = "project.lua"

type Config struct {
	Listen  string
	Editing bool
	Data    string
	Root    string
	Index   string
	DB      DB
	Limiter Limiter
	Env     *lua.LTable `lua:"-"`
}

type DB struct {
	Driver  string
	Conn    string
	SQLPath string
}

type Limiter struct {
	Enabled bool
	Max int
	Dur int
}

func Parse(projectDir string) (c Config, err error) {
	L := lua.NewState()
	defer L.Close()
	lelib.OpenLib(L)

	if err = os.Chdir(projectDir); err != nil {
		return
	}
	file, err := os.Open(ProjectFileName)
	if err != nil {
		return
	}
	str, err := io.ReadAll(file)
	if err != nil {
		return
	}
	if err = L.DoString(string(str)); err != nil {
		return
	}
	t := L.CheckTable(1)
	mapper := gluamapper.NewMapper(gluamapper.Option{TagName: "lua"})
	if err = mapper.Map(t, &c); err != nil {
		return
	}
	if env, ok := t.RawGetString("env").(*lua.LTable); ok {
		c.Env = env
	}
	return
}
