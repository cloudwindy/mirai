package leapp

import (
	lua "github.com/yuin/gopher-lua"
)

// read-only iterable object
// func objIterable(L *lua.LState, iterMap map[string]lua.LValue) lua.LValue {
// 	mt := map[string]lua.LGFunction{
// 		"__index": func(L *lua.LState) int {
// 			key := L.CheckString(2)
// 			if elem, ok := iterMap[key]; ok {
// 				L.Push(elem)
// 			} else {
// 				L.Push(lua.LNil)
// 			}
// 			return 1
// 		},
// 		"__call": func(L *lua.LState) int {
// 			cur := 0
// 			keys := make([]string, 0)
// 			for k := range iterMap {
// 				keys = append(keys, k)
// 			}
// 			closure := func(L *lua.LState) int {
// 				if cur >= len(iterMap) {
// 					L.Push(lua.LNil)
// 					return 1
// 				}
// 				k := keys[cur]
// 				L.Push(lua.LString(k))
// 				L.Push(iterMap[k])
// 				cur += 1
// 				return 2
// 			}
// 			L.Push(L.NewFunction(closure))
// 			return 1
// 		},
// 	}
// 	return objProxyFuncs(L, mt)
// }

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