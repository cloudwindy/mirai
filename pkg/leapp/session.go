package leapp

import (
	"time"

	"mirai/pkg/lazysess"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type Session struct {
	lazysess.Session
	index map[string]lua.LValue
}

func NewSession(L *lua.LState, s lazysess.Session) lua.LValue {
	sess := new(Session)
	sess.Session = s
	index := map[string]lua.LGFunction{
		"keys":  sessKeys,
		"save":  sessSave,
		"destroy": sessDestroy,
	}

	sess.index = make(map[string]lua.LValue)
	for i, v := range index {
		sess.index[i] = L.NewFunction(v)
	}

	indexFunc := L.NewFunction(sessIndex)
	newIndex := L.NewFunction(sessNewIndex)

	return objProxy(L, sess, indexFunc, newIndex)
}

func sessIndex(L *lua.LState) int {
	s := checkSess(L, 1)
	key := L.CheckString(2)
	if v, ok := s.index[key]; ok {
		L.Push(v)
	} else {
		value := s.Get(key)
		L.Push(luar.New(L, value))
	}
	return 1
}

func sessNewIndex(L *lua.LState) int {
	s := checkSess(L, 1)
	key := L.CheckString(2)
	if L.Get(3) == lua.LNil {
		s.Delete(key)
		return 0
	}
	value := L.Get(4)
	goval := gluamapper.ToGoValue(value, gluamapper.Option{
		NameFunc: func(s string) string { return s },
	})
	s.Set(key, goval)
	return 0
}

func sessKeys(L *lua.LState) int {
	s := checkSess(L, 1)
	k := L.NewTable()
	for _, key := range s.Keys() {
		k.Append(lua.LString(key))
	}
	L.Push(k)
	return 1
}

func sessSave(L *lua.LState) int {
	s := checkSess(L, 1)
	t := L.ToNumber(2)
	if t != 0 {
		s.SetExpiry(time.Duration(t) * time.Hour)
	}
	if err := s.Save(); err != nil {
		L.RaiseError("session save failed: %v", err)
	}
	return 0
}

func sessDestroy(L *lua.LState) int {
	s := checkSess(L, 1)
	if err := s.Destroy(); err != nil {
		L.RaiseError("session clear failed: %v", err)
	}
	return 0
}

func checkSess(L *lua.LState, n int) *Session {
	ud := L.CheckUserData(n)
	sess, ok := ud.Value.(*Session)
	if !ok {
		L.ArgError(n, "expected type Session")
	}
	return sess
}
