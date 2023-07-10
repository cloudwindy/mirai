package leapp

import (
	"mirai/pkg/luaengine"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
)

// Constants
const (
	LTApplication = "Application"
	methodAll     = "ALL"
)

// Listener hook
type Listener func(child *fiber.App) error

type Application struct {
	*fiber.App
	store  *session.Store
	listen Listener
}

// New creates a new instance of the Lua engine factory.
func New(store *session.Store, listen Listener) luaengine.Factory {
	return func(L *lua.LState) lua.LValue {
		// Create a new Fiber app
		app := new(Application)
		app.App = fiber.New()
		app.store = store
		app.listen = listen

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
			"listen":  appListen,
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

// appUse returns a Lua function that adds middleware to the app.
func appUse(L *lua.LState) int {
	app := checkApp(L, 1)
	var values []interface{}
	for i := 2; i < L.GetTop()+1; i++ {
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

// appAdd returns a Lua function that adds a route to the app.
func appAdd(L *lua.LState) int {
	app := checkApp(L, 1)
	method := L.CheckString(2)
	path := L.CheckString(3)
	handler := L.CheckFunction(4)
	app.Add(method, path, appHandler(L, app, handler))
	return 0
}

// appAddMethod returns a Lua function that adds a route with specified method to the app.
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

// appListen returns a Lua function that starts the app's listener.
func appListen(L *lua.LState) int {
	app := checkApp(L, 1)
	if err := app.listen(app.App); err != nil {
		L.RaiseError("listen: %v", err)
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
