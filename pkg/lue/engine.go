package lue

import (
	"context"
	"sync"

	"github.com/cloudwindy/mirai/lib"
	"github.com/cloudwindy/mirai/pkg/dir"
	lutpool "github.com/cloudwindy/mirai/pkg/lut/pool"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
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
	L      *lua.LState
	env    *lua.LTable
	mods   map[string]Module
	parent *Engine
	err    error
	lsp    *lutpool.LSPool
	sync.Mutex
}

type (
	Module func(E *Engine) lua.LValue
	Fun    func(E *Engine) int
)

func New(env map[string]any) *Engine {
	L := lua.NewState(lua.Options{
		SkipOpenLibs: true,
	})
	lib.Open(L)
	e := &Engine{
		L:    L,
		mods: make(map[string]Module),
	}
	e.lsp = lutpool.New(lua.Options{
		CallStackSize:   64,
		RegistrySize:    1024,
		RegistryMaxSize: 1024 * 10,
		SkipOpenLibs:    true,
	})
	if env != nil {
		t := L.NewTable()
		for k, v := range env {
			t.RawSetString(k, luar.New(L, v))
		}
		L.SetGlobal("env", t)
		e.env = t
	}
	L.SetGlobal("cmd", L.NewFunction(cmd))
	return e
}

func cmd(L *lua.LState) int {
	os := L.RegisterModule(lua.OsLibName, nil)
	execute := L.GetField(os, "execute")
	L.Insert(execute, 1)
	L.Call(L.GetTop()-1, 1)
	return 1
}

// Create a child engine for use in a different goroutine
func (e *Engine) New() (E *Engine, new bool) {
	if e.parent != nil {
		panic("Cannot create an engine from child")
	}
	L, new := e.lsp.Get()
	E = &Engine{
		L:      L,
		env:    e.env,
		mods:   e.mods,
		lsp:    e.lsp,
		parent: e,
	}
	if new {
		lib.Open(L)
		L.SetGlobal("env", e.env)
		for name, module := range e.mods {
			L.SetGlobal(name, module(E))
		}
	}
	return
}

func (e *Engine) Close() {
	if e.parent != nil {
		e.lsp.Put(e.L)
		return
	}
	e.lsp.Close()
	e.L.Close()
}

func (e *Engine) Register(name string, module Module) *Engine {
	e.Lock()
	defer e.Unlock()
	if e.parent == nil {
		e.mods[name] = module
	}
	e.L.SetGlobal(name, module(e))
	return e
}

func (e *Engine) Run(path string) *Engine {
	if e.parent != nil {
		panic("Cannot run a child engine.")
	}
	if e.err != nil {
		return e
	}
	e.Lock()
	defer e.Unlock()
	file, _, err := dir.Index(path, DefaultIndex)
	if err != nil {
		e.err = err
		return e
	}
	if err = e.L.DoFile(file); err != nil {
		e.err = err
		return e
	}
	return e
}

func (e *Engine) Eval(str string) *Engine {
	if e.parent != nil {
		panic("Cannot run a child engine.")
	}
	if e.err != nil {
		return e
	}
	if err := e.L.DoString(str); err != nil {
		e.err = err
		return e
	}
	return e
}

func (e *Engine) EvalWithContext(ctx context.Context, str string) *Engine {
	if e.parent != nil {
		panic("Cannot run a child engine.")
	}
	if e.err != nil {
		return e
	}
	e.L.SetContext(ctx)
	if err := e.L.DoString(str); err != nil {
		e.err = err
		return e
	}
	return e
}

func (e *Engine) Err() error {
	return e.err
}
