package leapp

import (
	lua "github.com/yuin/gopher-lua"
)

// for fasthttp

type (
	mtHttpGetFunc func(key string, defaultValue ...string) string
	mtHttpSetFunc func(key string, value string)
)

func mtHttpGetter(get mtHttpGetFunc) lua.LGFunction {
	return func(L *lua.LState) int {
		key := L.ToString(2)
		value := get(key)
		L.Push(lua.LString(value))
		return 1
	}
}

func mtHttpSetter(set mtHttpSetFunc) lua.LGFunction {
	return func(L *lua.LState) int {
		key := L.ToString(2)
		value := L.ToString(3)
		set(key, value)
		return 0
	}
}

// for generic key/value store

type (
	mtGetFunc func(key string) lua.LValue
	mtSetFunc func(key string, value lua.LValue)
)

func mtGetter(get mtGetFunc) lua.LGFunction {
	return func(L *lua.LState) int {
		key := L.ToString(2)
		value := get(key)
		L.Push(value)
		return 1
	}
}

func mtSetter(set mtSetFunc) lua.LGFunction {
	return func(L *lua.LState) int {
		key := L.ToString(2)
		value := L.Get(3)
		set(key, value)
		return 0
	}
}