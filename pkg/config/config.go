package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/cloudwindy/mirai/lib"
	"github.com/joho/godotenv"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

var ProjectFileName = "project.lua"

type Config struct {
	Listen   string
	Editing  bool
	DataPath string
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

func IsProject(projectDir string) (ok bool, err error) {
	_, err = os.Stat(path.Join(projectDir, ProjectFileName))
	if err == nil {
		ok = true
	} else if errors.Is(err, os.ErrNotExist) {
		err = nil
	}
	return
}

func Parse(projectDir string) (c Config, err error) {
	L := lua.NewState()
	defer L.Close()
	lib.Open(L)

	if err = os.Chdir(projectDir); err != nil {
		return
	}
	if err = L.DoFile(ProjectFileName); err != nil {
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

	env, err := godotenv.Read()
	if errors.Is(err, os.ErrNotExist) {
		err = nil
	}
	if err != nil {
		return
	}
	for k, v := range env {
		c.Env[k] = v
	}
	for _, rawEnv := range os.Environ() {
		k, v, ok := strings.Cut(rawEnv, "=")
		if !ok {
			panic(fmt.Sprintf("invalid environment variable: %s", rawEnv))
		}
		env[k] = v
	}
	for k, v := range env {
		switch k {
		case "INDEX":
			c.Index = v
		case "ROOT":
			c.Root = v
		case "LISTEN":
			c.Listen = v
		}
	}
	env = map[string]string{
		"INDEX":    c.Index,
		"ROOT":     c.Root,
		"LISTEN":   c.Listen,
		"DATAPATH": c.DataPath,
		"SQLPATH":  c.DB.SQLPath,
	}
	for k, v := range env {
		c.Env[k] = v
	}

	// defaults
	if c.Index == "" {
		c.Index = "."
	}
	if c.Listen == "" {
		c.Listen = ":80"
	}
	if c.DB.Driver == "" && c.DB.Conn == "" {
		c.DB = DB{
			Driver: "sqlite3",
			Conn:   ":memory:",
		}
	}

	return
}
