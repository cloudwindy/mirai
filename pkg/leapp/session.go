package leapp

import (
	"time"

	"github.com/cloudwindy/mirai/pkg/lazysess"
	"github.com/cloudwindy/mirai/pkg/lue"
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
	sess.index = E.MapFuncs(sessExports)

	index := E.LFun(sessIndex)
	newIndex := E.LFun(sessNewIndex)

	return E.Anonymous(sess, index, newIndex)
}

var sessExports = map[string]lue.Fun{
	"keys":    sessKeys,
	"save":    sessSave,
	"destroy": sessDestroy,
}

func sessIndex(E *lue.Engine) int {
	s := E.Data(1).(*Session)
	key := E.String(2)
	if v, ok := s.index[key]; ok {
		E.Push(v)
	} else {
		value := s.Get(key)
		E.Push(luar.New(E.L, value))
	}
	return 1
}

func sessNewIndex(E *lue.Engine) int {
	s := E.Data(1).(*Session)
	key := E.String(2)
	if E.Get(3) == lua.LNil {
		s.Delete(key)
		return 0
	}
	value := E.Get(4)
	goval := gluamapper.ToGoValue(value, gluamapper.Option{
		NameFunc: func(s string) string { return s },
	})
	s.Set(key, goval)
	return 0
}

func sessKeys(E *lue.Engine) int {
	s := E.Data(1).(*Session)
	k := E.NewTable()
	for _, key := range s.Keys() {
		k.Append(lua.LString(key))
	}
	E.Push(k)
	return 1
}

func sessSave(E *lue.Engine) int {
	s := E.Data(1).(*Session)
	if E.Top() > 1 {
		t := E.Number(2)
		s.SetExpiry(time.Duration(t) * time.Hour)
	}
	if err := s.Save(); err != nil {
		E.Error("session save: %v", err)
	}
	return 0
}

func sessDestroy(E *lue.Engine) int {
	s := E.Data(1).(*Session)
	if err := s.Destroy(); err != nil {
		E.Error("session clear: %v", err)
	}
	return 0
}
