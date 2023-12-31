package leapp

import (
	"strconv"
	"strings"

	"github.com/cloudwindy/mirai/pkg/lazysess"
	"github.com/cloudwindy/mirai/pkg/lue"
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
func NewContext(E *lue.Engine, app *Application, fc *fiber.Ctx) lua.LValue {
	c := new(Context)
	c.Ctx = fc
	c.store = app.c.Store

	index := E.NewTable()

	E.SetDict(index, map[string]string{
		"id":     c.Locals("requestid").(string),
		"method": c.Method(),
		"url":    ctxUrl(c),
		"path":   c.Path(),
		"body":   string(c.Body()),
	})
	E.SetFields(index, map[string]lua.LValue{
		"headers": ctxHeaders(E, c),
		"params":  ctxParams(E, c),
		"cookies": ctxCookies(E, c),
		"query":   ctxQuery(E, c),
		"state":   ctxState(E, c),
		"sess":    ctxSession(E, c),
		"form":    ctxForm(E, c),
	})
	E.SetFuncs(index, ctxExports)

	return E.Anonymous(c, index)
}

var ctxExports = map[string]lue.Fun{
	"type":  ctxType,
	"send":  ctxSend,
	"redir": ctxRedir,
	"next":  ctxNext,
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

func ctxHeaders(E *lue.Engine, c *Context) lua.LValue {
	return E.ReadWrite(mtHttpGetter(c.Get), mtHttpSetter(c.Set))
}

func ctxParams(E *lue.Engine, c *Context) lua.LValue {
	return E.SetDict(nil, c.AllParams())
}

func ctxCookies(E *lue.Engine, c *Context) lua.LValue {
	return NewCookies(E, c.Ctx)
}

func ctxQuery(E *lue.Engine, c *Context) lua.LValue {
	query := E.NewTable()
	E.SetDict(query, c.Queries())
	return query
}

func ctxForm(E *lue.Engine, c *Context) lua.LValue {
	return E.ReadOnly(mtHttpGetter(c.FormValue))
}

func ctxState(E *lue.Engine, c *Context) lua.LValue {
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
	return E.ReadWrite(mtGetter(getter), mtSetter(setter))
}

func ctxSession(E *lue.Engine, c *Context) lua.LValue {
	s := lazysess.New(c.Ctx, c.store)
	return NewSession(E, s)
}

func ctxType(E *lue.Engine) int {
	c := E.Data(1).(*Context)
	if E.Top() > 2 {
		c.Type(E.String(2), E.String(3))
	} else {
		c.Type(E.String(2))
	}
	return 0
}

func ctxSend(E *lue.Engine) int {
	var bodyLua lua.LValue
	c := E.Data(1).(*Context)
	status := 200
	bodyStr := ""
	if E.Top() > 2 {
		status = E.Int(2)
		bodyLua = E.Get(3)
	} else {
		bodyLua = E.Get(2)
	}
	switch body := bodyLua.(type) {
	case lua.LString:
		bodyStr = string(body)
	case lua.LNumber:
		bodyStr = strconv.Itoa(int(body))
	case *lua.LTable:
		bodyBytes, err := json.ValueEncode(body)
		if err != nil {
			E.Error("http send json: %v", err)
		}
		bodyStr = string(bodyBytes)
	default:
		E.Error("http send: unexpected type %s", body.Type().String())
	}
	if err := c.Status(status).SendString(bodyStr); err != nil {
		E.Error("http send: %v", err)
	}
	return 0
}

func ctxRedir(E *lue.Engine) int {
	c := E.Data(1).(*Context)
	status := 302
	loc := ""
	if E.Top() > 2 {
		status = E.Int(2)
		loc = E.String(3)
	} else {
		loc = E.String(2)
	}
	c.Status(status).Location(loc)
	return 0
}

func ctxNext(E *lue.Engine) int {
	c := E.Data(1).(*Context)
	if err := c.Next(); err != nil {
		E.Error("next: %v", err)
	}
	return 0
}
