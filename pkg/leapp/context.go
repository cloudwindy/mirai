package leapp

import (
	"strings"

	"mirai/pkg/lazysess"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/vadv/gopher-lua-libs/json"
	lua "github.com/yuin/gopher-lua"
)

type Context struct {
	*fiber.Ctx
	store *session.Store
}

// NewContext creates a new Lua table representing the Fiber context.
func NewContext(L *lua.LState, app *Application, fc *fiber.Ctx) lua.LValue {
	c := new(Context)
	c.Ctx = fc
	c.store = app.store

	index := L.NewTable()

	// Add string fields to the context table
	strDict := map[string]string{
		"id":     c.Locals("requestid").(string),
		"method": c.Method(),
		"url":    ctxUrl(c),
		"path":   c.Path(),
		"body":   string(c.Body()),
	}
	for k, v := range strDict {
		index.RawSetString(k, lua.LString(v))
	}

	// Add object fields to the context table
	objDict := map[string]lua.LValue{
		"headers": ctxHeaders(L, c),
		"params":  ctxParams(L, c),
		"cookies": ctxCookies(L, c),
		"query":   ctxQuery(L, c),
		"state":   ctxState(L, c),
		"sess":    ctxSession(L, c),
	}
	for k, v := range objDict {
		index.RawSetString(k, v)
	}

	// Add function fields to the context table
	funcDict := map[string]lua.LGFunction{
		"type":  ctxType,
		"send":  ctxSend,
		"redir": ctxRedir,
		"next":  ctxNext,
	}
	L.SetFuncs(index, funcDict)

	return objProxy(L, c, index)
}

func ctxUrl(c *Context) string {
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
	return u.String()
}

// ctxHeaders returns a Lua table representing the context's headers.
func ctxHeaders(L *lua.LState, c *Context) lua.LValue {
	return objReadWrite(L, mtHttpGetter(c.Get), mtHttpSetter(c.Set))
}

// ctxParams returns a Lua table representing the context's route parameters.
func ctxParams(L *lua.LState, c *Context) lua.LValue {
	params := L.NewTable()
	for k, v := range c.AllParams() {
		params.RawSetString(k, lua.LString(v))
	}
	return params
}

func ctxCookies(L *lua.LState, c *Context) lua.LValue {
	return NewCookies(L, c.Ctx)
}

// ctxQuery returns a Lua table representing the context's query parameters.
func ctxQuery(L *lua.LState, c *Context) lua.LValue {
	query := L.NewTable()
	q := c.Context().QueryArgs()
	q.VisitAll(func(key, value []byte) {
		query.RawSetString(string(key), lua.LString(value))
	})
	return query
}

// ctxState returns a Lua table representing the context's state.
func ctxState(L *lua.LState, c *Context) lua.LValue {
	getter := func(key string) lua.LValue {
		v, ok := c.Locals(key).(lua.LValue)
		if !ok {
			return lua.LNil
		}
		return v
	}
	setter := func(key string, value lua.LValue) {
		c.Locals(key, value)
	}
	return objReadWrite(L, mtGetter(getter), mtSetter(setter))
}

func ctxSession(L *lua.LState, c *Context) lua.LValue {
	s := lazysess.New(c.Ctx, c.store)
	return NewSession(L, s)
}

func ctxType(L *lua.LState) int {
	c := checkCtx(L, 1)
	if L.GetTop() > 2 {
		c.Type(L.CheckString(2), L.CheckString(3))
	} else {
		c.Type(L.CheckString(2))
	}
	return 0
}

func ctxSend(L *lua.LState) int {
	var bodyLua lua.LValue
	c := checkCtx(L, 1)
	status := 200
	bodyStr := ""
	if L.GetTop() > 2 {
		status = L.CheckInt(2)
		bodyLua = L.CheckAny(3)
	} else {
		bodyLua = L.CheckAny(2)
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
}

func ctxRedir(L *lua.LState) int {
	c := checkCtx(L, 1)
	status := 302
	loc := ""
	if L.GetTop() > 2 {
		status = L.CheckInt(2)
		loc = L.CheckString(3)
	} else {
		loc = L.CheckString(2)
	}
	c.Status(status).Location(loc)
	return 0
}

func ctxNext(L *lua.LState) int {
	c := checkCtx(L, 1)
	if err := c.Next(); err != nil {
		L.RaiseError("next: %v", err)
	}
	return 0
}

func checkCtx(L *lua.LState, n int) *Context {
	ud := L.CheckUserData(n)
	ctx, ok := ud.Value.(*Context)
	if !ok {
		L.ArgError(n, "expected type Context")
	}
	return ctx
}
