package leapp

import (
	"time"

	"mirai/pkg/lazysess"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func NewSession(L *lua.LState, s lazysess.Session) lua.LValue {
	sess := L.NewTable()

	funcs, mtFuncs := sessionFuncs(s)
	L.SetFuncs(sess, funcs)

	mt := L.NewTable()
	L.SetMetatable(sess, mt)
	L.SetFuncs(mt, mtFuncs)

	return sess
}

func sessionFuncs(s lazysess.Session) (map[string]lua.LGFunction, map[string]lua.LGFunction) {
	return map[string]lua.LGFunction{
			// normal fields
			"keys": func(L *lua.LState) int {
				k := L.NewTable()
				for _, key := range s.Keys() {
					k.Append(lua.LString(key))
				}
				L.Push(k)
				return 1
			},
			"save": func(L *lua.LState) int {
				t := L.ToNumber(1)
				if t != 0 {
					s.SetExpiry(time.Duration(t) * time.Hour)
				}
				if err := s.Save(); err != nil {
					L.RaiseError("session save failed: %v", err)
				}
				return 0
			},
			"clear": func(L *lua.LState) int {
				if err := s.Destroy(); err != nil {
					L.RaiseError("session clear failed: %v", err)
				}
				return 0
			},
		}, map[string]lua.LGFunction{
			// metatable
			"__index": func(L *lua.LState) int {
				key := L.CheckString(2)
				value := s.Get(key)
				L.Push(luar.New(L, value))
				return 1
			},
			"__newindex": func(L *lua.LState) int {
				key := L.CheckString(2)
				if L.Get(3) == lua.LNil {
					s.Delete(key)
					return 0
				}
				value := L.Get(3)
				goval := gluamapper.ToGoValue(value, gluamapper.Option{
					NameFunc: func(s string) string { return s },
				})
				s.Set(key, goval)
				return 0
			},
		}
}
