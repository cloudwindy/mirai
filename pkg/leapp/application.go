package leapp

import (
	"strings"
	"time"

	"github.com/cloudwindy/mirai/pkg/lue"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
)

// Constants
const (
	methodAll     = "ALL"
	LTApplication = "Application"
)

type (
	StartAndReloadHandler func() error
	StopHandler           func(timeout time.Duration) error
)

type Config struct {
	App    fiber.Router
	Store  *session.Store
	Start  StartAndReloadHandler
	Reload StartAndReloadHandler
	Stop   StopHandler
}

type Application struct {
	c       Config
	sub     bool
	globals map[string]lua.LValue
	fiber.Router
}

// New creates a new instance of the Lua engine factory.
func New(c Config) lue.Module {
	return func(E *lue.Engine) lua.LValue {
		// Create a new Fiber app
		app := new(Application)
		app.Router = c.App
		app.globals = make(map[string]lua.LValue)
		app.c = c

		// Set up the app functions
		index := E.L.NewTable()
		E.SetFuncs(index, map[string]lue.Fun{
			"start":   appStart,
			"reload":  appReload,
			"stop":    appStop,
			"set":     appSet,
			"sub":     appSub,
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
		})

		return E.Class(LTApplication, app, index)
	}
}

func appHandlerAsync(E *lue.Engine, app *Application, fn *lua.LFunction) fiber.Handler {
	// cannot pass upvalues to function
	if len(fn.Upvalues) > 0 {
		values := []string{}
		for _, v := range fn.Upvalues {
			values = append(values, v.Value().String())
		}
		E.L.RaiseError("cannot use passed values: %s", strings.Join(values, ", "))
	}
	return func(c *fiber.Ctx) error {
		E, _ := E.New()
		defer E.Close()
		env := E.L.CheckTable(lua.EnvironIndex)
		for k, v := range app.globals {
			env.RawSetString(k, v)
		}
		if err := E.CallLFun(fn, 1, env, NewContext(E, app, c)); err != nil {
			return errWithStackTrace(err, c)
		}
		return nil
	}
}

func appSub(E *lue.Engine) int {
	L := E.L
	app := E.Data(1).(*Application)
	prefix := L.CheckString(2)
	subapp := new(Application)
	subapp.Router = app.Group(prefix)
	subapp.sub = true
	L.Push(E.Class(LTApplication, subapp))
	return 1
}

// appUse adds middleware to the app.
func appUse(E *lue.Engine) int {
	L := E.L
	app := E.Data(1).(*Application)
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
			values = append(values, appHandlerAsync(E, app, val))
		}
	}
	app.Use(values...)
	return 0
}

// appAdd adds a route to the app.
func appAdd(E *lue.Engine) int {
	L := E.L
	app := E.Data(1).(*Application)
	method := L.CheckString(2)
	path := L.CheckString(3)
	handler := L.CheckFunction(4)
	app.Add(method, path, appHandlerAsync(E, app, handler))
	return 0
}

// appAddMethod adds a route with specified method to the app.
func appAddMethod(method string) lue.Fun {
	return func(E *lue.Engine) int {
		L := E.L
		app := E.Data(1).(*Application)
		path := L.CheckString(2)
		handler := L.CheckFunction(3)
		if method == methodAll {
			app.All(path, appHandlerAsync(E, app, handler))
		} else {
			app.Add(method, path, appHandlerAsync(E, app, handler))
		}
		return 0
	}
}

// appStart starts the app's listener.
func appStart(E *lue.Engine) int {
	app := E.Data(1).(*Application)
	if app.sub {
		E.L.RaiseError("app start: cannot start a subrouter")
	}
	if err := app.c.Start(); err != nil {
		E.L.RaiseError("app start: %v", err)
	}
	return 0
}

// appStart starts the app's listener.
func appReload(E *lue.Engine) int {
	app := E.Data(1).(*Application)
	if app.sub {
		E.L.RaiseError("app reload: cannot reload a subrouter")
	}
	if err := app.c.Reload(); err != nil {
		E.L.RaiseError("app reload: %v", err)
	}
	return 0
}

func appStop(E *lue.Engine) int {
	L := E.L
	app := E.Data(1).(*Application)
	if app.sub {
		E.L.RaiseError("app stop: cannot stop a subrouter")
	}
	timeout := float64(L.ToNumber(2))
	sec := float64(time.Second)
	if err := app.c.Stop(time.Duration(timeout * sec)); err != nil {
		L.RaiseError("app stop: %v", err)
	}
	return 0
}

func appSet(E *lue.Engine) int {
	L := E.L
	app := E.Data(1).(*Application)
	if app.sub {
		E.L.RaiseError("app set: cannot set a subrouter")
	}
	key := L.CheckString(2)
	value := L.CheckAny(3)
	app.globals[key] = value
	return 0
}

// errWithStackTrace adds a stack trace to the error if it's a Lua error.
func errWithStackTrace(e error, c *fiber.Ctx) error {
	if lerr, ok := e.(*lua.ApiError); ok {
		c.Locals("stacktrace", lerr.StackTrace)
		return errors.New(lerr.Object.String())
	}
	return e
}
