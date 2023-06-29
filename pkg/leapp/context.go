package leapp

import (
	"strings"

	"mirai/pkg/lazysess"

	"github.com/gofiber/fiber/v2"
	"github.com/vadv/gopher-lua-libs/json"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

func NewContext(L *lua.LState, c *fiber.Ctx) lua.LValue {
	ctx := L.NewTable()

	strDict := ctxFields(L, c)
	for k, v := range strDict {
		ctx.RawSetString(k, lua.LString(v))
	}
	objDict := map[string]lua.LValue{
		"headers": ctxHeaders(L, c),
		"params":  ctxParams(L, c),
		"cookies": ctxCookies(L, c),
		"query":   ctxQuery(L, c),
		"state":   ctxState(L, c),
	}
	if sess, ok := c.Locals(localSession).(lazysess.Session); ok {
		objDict["sess"] = NewSession(L, sess)
	}
	for k, v := range objDict {
		ctx.RawSetString(k, v)
	}
	L.SetFuncs(ctx, ctxFuncs(c))

	return ctx
}

func ctxFields(L *lua.LState, c *fiber.Ctx) map[string]string {
	u := new(strings.Builder)
	u.WriteString(c.Protocol())
	u.WriteString("://")
	u.WriteString(c.Hostname())
	u.WriteString(c.Path())
	q := c.Context().QueryArgs()
	if q.Len() > 0 {
		u.WriteByte('?')
		u.Write(q.QueryString())
	}

	return map[string]string{
		"id":     c.Locals("requestid").(string),
		"method": c.Method(),
		"url":    u.String(),
		"path":   c.Path(),
		"body":   string(c.Body()),
	}
}

func ctxHeaders(L *lua.LState, c *fiber.Ctx) lua.LValue {
	return objReadWrite(L, mtHttpGetter(c.Get), mtHttpSetter(c.Set))
}

func ctxParams(L *lua.LState, c *fiber.Ctx) lua.LValue {
	params := L.NewTable()
	for k, v := range c.AllParams() {
		params.RawSetString(k, lua.LString(v))
	}
	return params
}

func ctxCookies(L *lua.LState, c *fiber.Ctx) lua.LValue {
	mapper := gluamapper.NewMapper(gluamapper.Option{
		TagName: "json",
	})
	setter := func(L *lua.LState) int {
		name := L.CheckString(2)
		t := L.CheckTable(3)
		ck := new(fiber.Cookie)
		if err := mapper.Map(t, ck); err != nil {
			L.RaiseError("cookies set: %v", err)
		}
		ck.Name = name
		c.Cookie(ck)
		return 0
	}
	return objReadWrite(L, mtHttpGetter(c.Cookies), setter)
}

func ctxQuery(L *lua.LState, c *fiber.Ctx) lua.LValue {
	query := L.NewTable()
	q := c.Context().QueryArgs()
	q.VisitAll(func(key, value []byte) {
		query.RawSetString(string(key), lua.LString(value))
	})
	return query
}

func ctxState(L *lua.LState, c *fiber.Ctx) lua.LValue {
	s := c.Locals(localState).(state)
	getter := func(key string) lua.LValue {
		if v, ok := s[key]; ok {
			return v
		}
		return lua.LNil
	}
	setter := func(key string, value lua.LValue) {
		s[key] = value
	}
	return objReadWrite(L, mtGetter(getter), mtSetter(setter))
}

func ctxFuncs(c *fiber.Ctx) map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"type": func(L *lua.LState) int {
			if L.GetTop() > 1 {
				c.Type(L.CheckString(1), L.CheckString(2))
			} else {
				c.Type(L.CheckString(1))
			}
			return 0
		},
		"send": func(L *lua.LState) int {
			var (
				status  = 200
				bodyLua lua.LValue
				bodyStr string
			)
			if L.GetTop() > 1 {
				status = L.CheckInt(1)
				bodyLua = L.CheckAny(2)
			} else {
				bodyLua = L.CheckAny(1)
			}
			switch body := bodyLua.(type) {
			case lua.LString:
				bodyStr = string(body)
			case lua.LNumber:
				status = int(body)
			case *lua.LTable:
				bodyBytes, err := json.ValueEncode(body)
				if err != nil {
					L.RaiseError("http send json failed: %v", err)
				}
				bodyStr = string(bodyBytes)
			default:
				L.RaiseError("http send failed: unexpected type %s", body.Type().String())
			}
			if err := c.Status(status).SendString(bodyStr); err != nil {
				L.RaiseError("http send failed: %v", err)
			}
			return 0
		},
		"redir": func(L *lua.LState) int {
			status := 302
			loc := ""
			if L.GetTop() > 1 {
				status = L.CheckInt(1)
				loc = L.CheckString(2)
			} else {
				loc = L.CheckString(1)
			}
			c.Status(status).Location(loc)
			return 0
		},
		"next": func(L *lua.LState) int {
			if err := c.Next(); err != nil {
				L.RaiseError("next: %v", err)
			}
			return 0
		},
	}
}
