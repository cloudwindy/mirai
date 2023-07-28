package leapp

import (
	"github.com/cloudwindy/mirai/pkg/lue"
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
	ck.index = E.MapFuncs(ckExports)

	indexFunc := E.LFun(ckIndex)
	return E.Anonymous(ck, indexFunc)
}

var ckExports = map[string]lue.Fun{
	"set":   ckSet,
	"clear": ckClear,
}

func ckIndex(E *lue.Engine) int {
	ck := E.Data(1).(*Cookies)
	key := E.String(2)
	if v, ok := ck.index[key]; ok {
		E.Push(v)
	} else {
		value := ck.Cookies(key)
		E.PushString(value)
	}
	return 1
}

func ckSet(E *lue.Engine) int {
	ck := E.Data(1).(*Cookies)
	key := E.String(2)
	if E.IsNil(3) {
		ck.ClearCookie(key)
		return 0
	}
	mapper := gluamapper.NewMapper(gluamapper.Option{
		TagName: "json",
	})
	value := E.String(4)
	options := E.Table(5)
	c := new(fiber.Cookie)
	if options != nil {
		if err := mapper.Map(options, ck); err != nil {
			E.Error("cookies set: %v", err)
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
