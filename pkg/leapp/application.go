package leapp

import (
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
	StartHandler  func(listen string) error
	ReloadHandler func() error
	StopHandler   func(timeout time.Duration) error
)

type Config struct {
	App    fiber.Router
	Store  *session.Store
	Start  StartHandler
	Reload ReloadHandler
	Stop   StopHandler
}

type Application struct {
	c   Config
	up  bool
	sub bool
	fiber.Router
}

// New creates a new instance of the Lua engine factory.
func New(c Config) lue.Module {
	return func(E *lue.Engine) lua.LValue {
		// Create a new Fiber app
		app := new(Application)
		app.Router = c.App
		app.c = c

		// Set up the app functions
		index := E.NewTable()
		E.SetFuncs(index, appExports)

		return E.Class(LTApplication, app, index)
	}
}

var appExports = map[string]lue.Fun{
	"start":   appStart,
	"reload":  appReload,
	"stop":    appStop,
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
}

func appHandlerAsync(E *lue.Engine, app *Application, fn *lua.LFunction) fiber.Handler {
	return func(c *fiber.Ctx) error {
		E, _ := E.New()
		defer E.Close()
		env := E.Table(lua.EnvironIndex)
		if err := E.CallLFun(fn, env, 0, NewContext(E, app, c)); err != nil {
			return errWithStackTrace(err, c)
		}
		return nil
	}
}

func appSub(E *lue.Engine) int {
	app := E.Data(1).(*Application)
	prefix := E.String(2)
	subapp := new(Application)
	subapp.Router = app.Group(prefix)
	subapp.sub = true
	E.Push(E.Class(LTApplication, subapp))
	return 1
}

// appUse adds middleware to the app.
func appUse(E *lue.Engine) int {
	app := E.Data(1).(*Application)
	var values []any
	for i := 2; i <= E.Top(); i++ {
		switch val := E.Get(i).(type) {
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
	app := E.Data(1).(*Application)
	method := E.String(2)
	path := E.String(3)
	handler := E.Fun(4)
	app.Add(method, path, appHandlerAsync(E, app, handler))
	return 0
}

// appAddMethod adds a route with specified method to the app.
func appAddMethod(method string) lue.Fun {
	return func(E *lue.Engine) int {
		app := E.Data(1).(*Application)
		path := E.String(2)
		handler := E.Fun(3)
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
	if app.up {
		E.Error("app start: already started")
	}
	if app.sub {
		E.Error("app start: cannot start a subrouter")
	}
	listen := ""
	if E.Top() > 1 {
		listen = E.String(2)
	}
	app.up = true
	if err := app.c.Start(listen); err != nil {
		E.Error("app start: %v", err)
	}
	return 0
}

// appStart starts the app's listener.
func appReload(E *lue.Engine) int {
	app := E.Data(1).(*Application)
	if app.sub {
		E.Error("app reload: cannot reload a subrouter")
	}
	if app.c.Reload == nil {
		E.Error("app reload: not supported")
	}
	if err := app.c.Reload(); err != nil {
		E.Error("app reload: %v", err)
	}
	return 0
}

func appStop(E *lue.Engine) int {
	app := E.Data(1).(*Application)
	if !app.up {
		E.Error("app stop: already stopped")
	}
	if app.sub {
		E.Error("app stop: cannot stop a subrouter")
	}
	timeout := float64(0)
	if E.Top() > 1 {
		timeout = E.Number(2)
	}
	app.up = false
	const s = float64(time.Second)
	if err := app.c.Stop(time.Duration(timeout * s)); err != nil {
		E.Error("app stop: %v", err)
	}
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
