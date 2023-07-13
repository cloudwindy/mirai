package lue

import (
	"sync"

	"mirai/pkg/dir"
	"mirai/pkg/lelib"
	lutpool "mirai/pkg/lut/pool"

	lua "github.com/yuin/gopher-lua"
)

// Package info
const (
	LuaVersion = lua.LuaVersion
)

// Package variables
var (
	DefaultIndex = "index.lua"
	lspool       = lutpool.New()
)

type Engine struct {
	L       *lua.LState
	Values  map[string]lua.LValue
	env     *lua.LTable
	path    string
	modules map[string]Module
	parent  *Engine
	sync.Mutex
}

type (
	Module func(E *Engine) lua.LValue
	Fun    func(E *Engine) int
)

func New(path string, env *lua.LTable) *Engine {
	L := lua.NewState()
	lelib.OpenLib(L)
	L.SetGlobal("env", env)
	return &Engine{L: L, path: path, modules: make(map[string]Module)}
}

// Create a child engine for use in a different goroutine
func (e *Engine) New() (E *Engine, new bool) {
	if e.parent != nil {
		panic("Cannot create an engine from child")
	}
	L, new := lspool.Get()
	E = &Engine{
		L:       L,
		env:     e.env,
		modules: e.modules,
		parent:  e,
	}
	if new {
		lelib.OpenLib(L)
		L.SetGlobal("env", e.env)
		for name, module := range e.modules {
			L.SetGlobal(name, module(E))
		}
	}
	return
}

func (e *Engine) Close() {
	if e.parent == nil {
		panic("Close must be called from a child")
	}
	lspool.Put(e.L)
}

func (e *Engine) MapFuncs(funs map[string]Fun) map[string]lua.LValue {
	dict := make(map[string]lua.LValue)
	for name, fun := range funs {
		dict[name] = e.LFun(fun)
	}
	return dict
}

func (e *Engine) SetFuncs(tb *lua.LTable, funs map[string]Fun) *lua.LTable {
	if tb == nil {
		tb = e.L.NewTable()
	}
	for k, v := range funs {
		tb.RawSetString(k, e.LFun(v))
	}
	return tb
}

func (e *Engine) SetFields(tb *lua.LTable, fields map[string]lua.LValue) *lua.LTable {
	if tb == nil {
		tb = e.L.NewTable()
	}
	for k, v := range fields {
		tb.RawSetString(k, v)
	}
	return tb
}

func (e *Engine) SetDict(tb *lua.LTable, dict map[string]string) *lua.LTable {
	if tb == nil {
		tb = e.L.NewTable()
	}
	for k, v := range dict {
		tb.RawSetString(k, lua.LString(v))
	}
	return tb
}

func (e *Engine) LFun(fn Fun) *lua.LFunction {
	return e.L.NewFunction(e.LGFun(fn))
}

func (e *Engine) LGFun(fn Fun) lua.LGFunction {
	return func(*lua.LState) int {
		return fn(e)
	}
}

// Get all arguments
func (e *Engine) Arguments() []lua.LValue {
	L := e.L
	params := make([]lua.LValue, 0, L.GetTop())
	for i := 1; i <= L.GetTop(); i++ {
		params = append(params, L.Get(i))
	}
	return params
}

func (e *Engine) Register(name string, module Module) *Engine {
	e.Lock()
	defer e.Unlock()
	e.modules[name] = module
	if e.parent != nil {
		e.parent.Lock()
		e.parent.L.SetGlobal(name, module(e.parent))
		e.parent.Unlock()
	}
	e.L.SetGlobal(name, module(e))
	return e
}

func (e *Engine) Run() (err error) {
	if e.parent != nil {
		panic("Cannot run a child engine.")
	}
	e.Lock()
	defer e.Unlock()
	file, _, err := dir.Index(e.path, DefaultIndex)
	if err != nil {
		return
	}
	if err = e.L.DoFile(file); err != nil {
		return
	}
	return
}
