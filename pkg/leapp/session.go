package leapp

import (
	"time"

	"mirai/pkg/lazysess"
	"mirai/pkg/lue"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type Session struct {
	lazysess.Session
	index map[string]lua.LValue
}

func NewSession(E *lue.Engine, s lazysess.Session) lua.LValue {
	sess := new(Session)
	sess.Session = s
	sess.index = E.MapFuncs(map[string]lue.Fun{
		"keys":    sessKeys,
		"save":    sessSave,
		"destroy": sessDestroy,
	})

	index := E.LFun(sessIndex)
	newIndex := E.LFun(sessNewIndex)

	return E.Anonymous(sess, index, newIndex)
}

func sessIndex(E *lue.Engine) int {
	L := E.L
	s := E.Data(1).(*Session)
	key := L.CheckString(2)
	if v, ok := s.index[key]; ok {
		L.Push(v)
	} else {
		value := s.Get(key)
		L.Push(luar.New(L, value))
	}
	return 1
}

func sessNewIndex(E *lue.Engine) int {
	L := E.L
	s := E.Data(1).(*Session)
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

func sessKeys(E *lue.Engine) int {
	L := E.L
	s := E.Data(1).(*Session)
	k := L.NewTable()
	for _, key := range s.Keys() {
		k.Append(lua.LString(key))
	}
	L.Push(k)
	return 1
}

func sessSave(E *lue.Engine) int {
	L := E.L
	s := E.Data(1).(*Session)
	t := L.ToNumber(2)
	if t != 0 {
		s.SetExpiry(time.Duration(t) * time.Hour)
	}
	if err := s.Save(); err != nil {
		L.RaiseError("session save failed: %v", err)
	}
	return 0
}

func sessDestroy(E *lue.Engine) int {
	L := E.L
	s := E.Data(1).(*Session)
	if err := s.Destroy(); err != nil {
		L.RaiseError("session clear failed: %v", err)
	}
	return 0
}
