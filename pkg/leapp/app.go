package leapp

import (
	"mirai/pkg/lazysess"
	"mirai/pkg/luaengine"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
)

// Constants
const (
	methodAll = "ALL"

	localState   = "luaengine_state"
	localSession = "luaengine_session"
)

type (
	state map[string]lua.LValue
	// listen hook
	Listen func(child *fiber.App) error
)

func New(store *session.Store, listen Listen) luaengine.Factory {
	return func(L *lua.LState) lua.LValue {
		a := fiber.New()
		app := L.NewTable()

		a.Use(func(c *fiber.Ctx) error {
			c.Locals(localState, make(state))
			sess := lazysess.New(c, store)
			c.Locals(localSession, sess)
			return c.Next()
		})
		L.SetFuncs(app, appFuncs(L, a, listen))
		return app
	}
}

func appFuncs(L *lua.LState, app *fiber.App, listen Listen) map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"use":     appUse(app),
		"all":     appAdd(app, methodAll),
		"get":     appAdd(app, fiber.MethodGet),
		"head":    appAdd(app, fiber.MethodHead),
		"post":    appAdd(app, fiber.MethodPost),
		"put":     appAdd(app, fiber.MethodPut),
		"delete":  appAdd(app, fiber.MethodDelete),
		"connect": appAdd(app, fiber.MethodConnect),
		"options": appAdd(app, fiber.MethodOptions),
		"trace":   appAdd(app, fiber.MethodTrace),
		"patch":   appAdd(app, fiber.MethodPatch),
		"upgrade": wsAppUpgrade(app),
		"listen":  appListen(app, listen),
	}
}

func appHandler(L *lua.LState, fn *lua.LFunction) fiber.Handler {
	p := lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}
	return func(c *fiber.Ctx) error {
		if err := L.CallByParam(p, NewContext(L, c)); err != nil {
			return errWithStackTrace(err, c)
		}
		if str := L.ToString(1); str != "" {
			return errors.New(str)
		}
		return nil
	}
}

func appUse(app *fiber.App) lua.LGFunction {
	return func(L *lua.LState) int {
		var values []interface{}
		for i := 1; i < L.GetTop(); i++ {
			switch val := L.CheckAny(i).(type) {
			case lua.LString:
				values = append(values, string(val))
			case *lua.LTable:
				var list []string
				val.ForEach(func(_, v lua.LValue) {
					if str := lua.LVAsString(v); str != "" {
						list = append(list, str)
					}
				})
				values = append(values, list)
			case *lua.LFunction:
				values = append(values, appHandler(L, val))
			}
		}
		app.Use(values...)
		return 0
	}
}

func appAdd(app *fiber.App, method string) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		handler := L.CheckFunction(2)
		if method == methodAll {
			app.All(path, appHandler(L, handler))
		} else {
			app.Add(method, path, appHandler(L, handler))
		}
		return 0
	}
}

func appListen(app *fiber.App, listen Listen) lua.LGFunction {
	return func(L *lua.LState) int {
		if err := listen(app); err != nil {
			L.RaiseError("listen: %v", err)
		}
		return 0
	}
}

// Lua error stacktrace helper.
func errWithStackTrace(e error, c *fiber.Ctx) error {
	if lerr, ok := e.(*lua.ApiError); ok {
		c.Locals("stacktrace", lerr.StackTrace)
		return errors.New(lerr.Object.String())
	}
	return e
}
