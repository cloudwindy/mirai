package leapp

import (
	"time"

	"mirai/pkg/luaengine"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
)

// Constants
const (
	methodAll     = "ALL"
)

// Start hook
type Start func(child *fiber.App) error

type Application struct {
	*fiber.App
	store *session.Store
	start Start
}

// New creates a new instance of the Lua engine factory.
func New(store *session.Store, start Start) luaengine.Factory {
	return func(L *lua.LState) lua.LValue {
		// Create a new Fiber app
		app := new(Application)
		app.App = fiber.New()
		app.store = store
		app.start = start

		// Set up the app functions
		index := L.NewTable()
		L.SetFuncs(index, map[string]lua.LGFunction{
			"use":     appUse,
			"add":     appAdd,
			"all":     appAddMethod(methodAll),
			"get":     appAddMethod(fiber.MethodGet),
			"head":    appAddMethod(fiber.MethodHead),
			"post":    appAddMethod(fiber.MethodPost),
			"put":     appAddMethod(fiber.MethodPut),
			"delete":  appAddMethod(fiber.MethodDelete),
			"connect": appAddMethod(fiber.MethodConnect),
			"options": appAddMethod(fiber.MethodOptions),
			"trace":   appAddMethod(fiber.MethodTrace),
			"patch":   appAddMethod(fiber.MethodPatch),
			"upgrade": wsAppUpgrade,
			"start":   appStart,
			"stop":    appStop,
		})

		return objProxy(L, app, index)
	}
}

// appHandler creates a Fiber handler function from a Lua function.
func appHandler(L *lua.LState, app *Application, fn *lua.LFunction) fiber.Handler {
	p := lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}
	return func(c *fiber.Ctx) error {
		if err := L.CallByParam(p, NewContext(L, app, c)); err != nil {
			return errWithStackTrace(err, c)
		}
		return nil
	}
}

// appUse adds middleware to the app.
func appUse(L *lua.LState) int {
	app := checkApp(L, 1)
	var values []interface{}
	for i := 2; i <= L.GetTop(); i++ {
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
			values = append(values, appHandler(L, app, val))
		}
	}
	app.Use(values...)
	return 0
}

// appAdd adds a route to the app.
func appAdd(L *lua.LState) int {
	app := checkApp(L, 1)
	method := L.CheckString(2)
	path := L.CheckString(3)
	handler := L.CheckFunction(4)
	app.Add(method, path, appHandler(L, app, handler))
	return 0
}

// appAddMethod adds a route with specified method to the app.
func appAddMethod(method string) lua.LGFunction {
	return func(L *lua.LState) int {
		app := checkApp(L, 1)
		path := L.CheckString(2)
		handler := L.CheckFunction(3)
		if method == methodAll {
			app.All(path, appHandler(L, app, handler))
		} else {
			app.Add(method, path, appHandler(L, app, handler))
		}
		return 0
	}
}

// appStart starts the app's listener.
func appStart(L *lua.LState) int {
	app := checkApp(L, 1)
	if err := app.start(app.App); err != nil {
		L.RaiseError("app start: %v", err)
	}
	return 0
}

func appStop(L *lua.LState) int {
	app := checkApp(L, 1)
	timeout := float64(L.ToNumber(2))
	sec := float64(time.Second)
	if timeout != 0 {
		if err := app.ShutdownWithTimeout(time.Duration(timeout * sec)); err != nil {
			L.RaiseError("app stop: %v", err)
		}
	} else {
		if err := app.Shutdown(); err != nil {
			L.RaiseError("app stop: %v", err)
		}
	}
	return 0
}

func checkApp(L *lua.LState, n int) *Application {
	ud := L.CheckUserData(n)
	app, ok := ud.Value.(*Application)
	if !ok {
		L.ArgError(n, "expected type Application")
	}
	return app
}

// errWithStackTrace adds a stack trace to the error if it's a Lua error.
func errWithStackTrace(e error, c *fiber.Ctx) error {
	if lerr, ok := e.(*lua.ApiError); ok {
		c.Locals("stacktrace", lerr.StackTrace)
		return errors.New(lerr.Object.String())
	}
	return e
}
