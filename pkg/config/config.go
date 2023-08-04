package config

import (
	"io"
	"os"

	"github.com/cloudwindy/mirai/lib"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

var ProjectFileName = "project.lua"

type Config struct {
	Listen   string
	Editing  bool
	Data     string
	Root     string
	Index    string
	Pid      string
	DB       DB
	Limiter  Limiter
	Commands map[string]string
	Env      map[string]any
}

type DB struct {
	Driver  string
	Conn    string
	SQLPath string
}

type Limiter struct {
	Enabled bool
	Max     int
	Dur     int
}

func Parse(projectDir string) (c Config, err error) {
	L := lua.NewState()
	defer L.Close()
	lib.Open(L)

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
	mapper := gluamapper.NewMapper(gluamapper.Option{
		TagName: "lua",
		NameFunc: func(s string) string {
			return s
		},
	})
	c.Env = make(map[string]any)
	if err = mapper.Map(t, &c); err != nil {
		return
	}
	if c.Index == "" {
		c.Index = "."
	}
	if c.DB.Driver == "" || c.DB.Conn == "" {
		c.DB = DB{
			Driver: "sqlite3",
			Conn:   ":memory:",
		}
	}
	env := map[string]string{
		"INDEX":    c.Index,
		"ROOT":     c.Root,
		"LISTEN":   c.Listen,
		"DATAPATH": c.Data,
		"SQLPATH":  c.DB.SQLPath,
	}
	for k, v := range env {
		c.Env[k] = v
	}
	return
}
