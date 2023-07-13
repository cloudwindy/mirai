package leapp

import (
	"mirai/pkg/lue"

	"github.com/gofiber/fiber/v2"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

type Cookies struct {
	*fiber.Ctx
	index map[string]lua.LValue
}

func NewCookies(E *lue.Engine, c *fiber.Ctx) lua.LValue {
	ck := new(Cookies)
	ck.Ctx = c
	ck.index = E.MapFuncs(map[string]lue.Fun{
		"set":   ckSet,
		"clear": ckClear,
	})

	indexFunc := E.LFun(ckIndex)
	return E.Anonymous(ck, indexFunc)
}

func ckIndex(E *lue.Engine) int {
	L := E.L
	ck := E.Data(1).(*Cookies)
	key := L.CheckString(2)
	if v, ok := ck.index[key]; ok {
		L.Push(v)
	} else {
		value := ck.Cookies(key)
		L.Push(lua.LString(value))
	}
	return 1
}

func ckSet(E *lue.Engine) int {
	L := E.L
	ck := E.Data(1).(*Cookies)
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
	if options != nil {
		if err := mapper.Map(options, ck); err != nil {
			L.RaiseError("cookies set: %v", err)
		}
	}
	c.Name = key
	c.Value = value
	ck.Cookie(c)
	return 0
}

func ckClear(E *lue.Engine) int {
	ck := E.Data(1).(*Cookies)
	ck.ClearCookie()
	return 0
}
