package leapp

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

type Cookies struct {
	*fiber.Ctx
	index map[string]lua.LValue
}

func NewCookies(L *lua.LState, c *fiber.Ctx) lua.LValue {
	ck := new(Cookies)
	ck.Ctx = c
	index := map[string]lua.LGFunction{
		"set":   ckSet,
		"clear": ckClear,
	}

	ck.index = make(map[string]lua.LValue)
	for i, v := range index {
		ck.index[i] = L.NewFunction(v)
	}

	indexFunc := L.NewFunction(ckIndex)
	return objAnonymous(L, ck, indexFunc)
}

func ckIndex(L *lua.LState) int {
	ck := checkCk(L, 1)
	key := L.CheckString(2)
	if v, ok := ck.index[key]; ok {
		L.Push(v)
	} else {
		value := ck.Cookies(key)
		L.Push(lua.LString(value))
	}
	return 1
}

func ckSet(L *lua.LState) int {
	ck := checkCk(L, 1)
	key := L.CheckString(2)
	if L.Get(3) == lua.LNil {
		ck.ClearCookie(key)
		return 0
	}
	mapper := gluamapper.NewMapper(gluamapper.Option{
		TagName: "json",
	})
	value := L.CheckString(4)
	options := L.ToTable(5)
	c := new(fiber.Cookie)
	if ck != nil {
		if err := mapper.Map(options, ck); err != nil {
			L.RaiseError("cookies set: %v", err)
		}
	}
	c.Name = key
	c.Value = value
	ck.Cookie(c)
	return 0
}

func ckClear(L *lua.LState) int {
	ck := checkCk(L, 1)
	if err := ck; err != nil {
		L.RaiseError("session clear failed: %v", err)
	}
	return 0
}

func checkCk(L *lua.LState, n int) *Cookies {
	ud := L.CheckUserData(n)
	ck, ok := ud.Value.(*Cookies)
	if !ok {
		L.ArgError(n, "expected type Cookies")
	}
	return ck
}
