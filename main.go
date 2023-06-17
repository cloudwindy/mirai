package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"mirai/pkg/config"
	"mirai/pkg/luapool"
	"mirai/pkg/luasql"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/session"
	sbolt "github.com/gofiber/storage/bbolt"
	"github.com/pkg/errors"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	"github.com/zs5460/art"
	luar "layeh.com/gopher-luar"
)

const (
	Version    = "1.0"
	ConfigFile = "./config.lua"
)

var (
	succ = color.New(color.FgGreen)
	info = color.New(color.FgBlue)
	warn = color.New(color.FgHiYellow)
)

func init() {
	lua.LuaPathDefault += ";./lib/?.lua;./lib/?/init.lua"
}

func main() {
	// 设置退出信号
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(term)
	wg := new(sync.WaitGroup)
	defer wg.Wait()

	// 优雅退出
	done := make(chan any, 1)
	go hardStop(term, done)

	color.Blue(art.String("Mirai Project"))
	fmt.Println(" Mirai Server " + Version + " with " + lua.LuaVersion)
	fmt.Println(" Fiber " + fiber.Version)

	c := config.Get(ConfigFile)
	defer luaPool.Shutdown()

	app := fiber.New(fiber.Config{
		ServerHeader:          "Mirai Server",
		DisableStartupMessage: true,
	})
	app.Use(PrintTimer("total", "Total Time"))
	app.Use(favicon.New())
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Done: func(c *fiber.Ctx, logString []byte) {
			if tb := c.Locals("stacktrace"); tb != nil {
				color.Red("%s", tb)
			}
		},
	}))
	app.Use(cors.New())
	app.Use(compress.New())
	app.Use(pprof.New())

	storage := sbolt.New(sbolt.Config{
		Database: path.Join(c.DataDir, "fiber.db"),
	})
	store := session.New(session.Config{
		Storage: storage,
	})

	api := app.Group("/api")
	api.Use(limiter.New(limiter.Config{
		Max:        200,
		Expiration: 10 * time.Minute,
	}))
	api.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))
	api.Use(PrintTimer("exec", "Script Execution"))
	api.All("/v1/:scriptname?/*", luaHandler(c.ScriptDir, c.Env, c.DB, store))

	admin := api.Group("/admin")
	if c.Editing {
		admin.All("/files/:path?", filesHandler(c.ScriptDir))
		fmt.Print(" Editing: ")
		warn.Println("Allowed")
	}

	app.Use(etag.New())
	app.Use(cache.New(cache.Config{
		Storage:      storage,
		CacheHeader:  "Cache-Status",
		CacheControl: false,
		Expiration:   72 * time.Hour,
	}))
	if c.RootDir != "" {
		app.Static("/", c.RootDir, fiber.Static{
			ByteRange: true,
		})
		app.Get("*", func(c *fiber.Ctx) error {
			c.Path("/")
			return c.RestartRouting()
		})
		fmt.Print(" Static: ")
		succ.Print("Serving")
		fmt.Printf(" from ")
		info.Println(c.RootDir)
	}

	fmt.Println()
	fmt.Print(" Listening at ")
	info.Println(c.Listen)
	if err := app.Listen(c.Listen); err != nil {
		panic(errors.Wrap(err, "无法启动 HTTP 服务器"))
	}
}

func StartTimer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		c.Locals("timer", start)
		return c.Next()
	}
}

func PrintTimer(name string, desc string, started ...bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var (
			start    time.Time
			bstarted = !(len(started) != 0 && started[0])
			err      error
		)
		if bstarted {
			start = time.Now()
			c.Locals("timer", start)
			err = c.Next()
		}
		stop := time.Now()
		if start.IsZero() {
			start = c.Locals("timer").(time.Time)
		}
		timing := new(strings.Builder)
		timing.WriteString(name)
		if len(desc) != 0 {
			timing.WriteString(";desc=")
			timing.WriteByte('"')
			timing.WriteString(desc)
			timing.WriteByte('"')
		}
		timing.WriteString(";dur=")
		timing.WriteString(fmt.Sprintf("%.02f", float64(stop.Sub(start).Microseconds())/1000))
		c.Append(fiber.HeaderServerTiming, timing.String())
		if !bstarted {
			err = c.Next()
		}
		return err
	}
}

// Global LState pool
var (
	filesStat = make(map[string]fs.FileInfo)
	luaCache  = make(map[string]*lua.FunctionProto)
	luaPool   = &luapool.StatePool{
		Saved: make([]*lua.LState, 0, 4),
	}
)

func luaHandler(base string, env *lua.LTable, db config.DB, store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file := c.Params("scriptname", "index") + ".lua"
		file = path.Join(base, file)

		stat, err := os.Stat(file)
		if os.IsNotExist(err) {
			return fiber.ErrNotFound
		}
		if err != nil {
			return err
		}

		prev, ok := filesStat[file]
		if !ok || stat.Size() != prev.Size() || stat.ModTime() != prev.ModTime() {
			proto, err := luapool.CompileLua(file)
			if err != nil {
				return err
			}
			luaCache[file] = proto
			filesStat[file] = stat
		}
		L := luaPool.Get()
		defer luaPool.Put(L)
		s, err := store.Get(c)
		if err != nil {
			return err
		}

		registerEnv(L, env)
		registerRequest(L, c)
		registerResponse(L, c)
		registerSession(L, s)
		registerDB(L, db)

		if err := luapool.DoCompiledFile(L, luaCache[file]); err != nil {
			if lerr := err.(*lua.ApiError); lerr != nil {
				c.Locals("stacktrace", lerr.StackTrace)
				return &apiError{msg: lerr.Object.String()}
			}
			return err
		}

		return nil
	}
}

func LogTime(t time.Time) {
	fmt.Printf("%.02fms\n", float64(time.Since(t).Microseconds())/1000)
}

type apiError struct {
	msg string
}

func (e *apiError) Error() string {
	return e.msg
}

func filesHandler(base string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file := c.Params("path")
		if file == "" {
			dir, err := os.ReadDir(base)
			if err != nil {
				return err
			}
			names := make([]string, 0)
			for _, file := range dir {
				names = append(names, file.Name())
			}
			return c.JSON(names)
		}
		file = path.Join(base, file)
		switch c.Method() {
		case "GET":
			return c.SendFile(file)
		case "PUT":
			if err := os.WriteFile(file, c.Body(), 0o644); err != nil {
				return err
			}
			return c.SendString("ok")
		case "DELETE":
			if err := os.Remove(file); err != nil {
				return err
			}
			return c.SendString("ok")
		}
		return nil
	}
}

func registerEnv(L *lua.LState, env *lua.LTable) {
	L.SetGlobal("env", env)
}

func registerRequest(L *lua.LState, c *fiber.Ctx) {
	req := L.NewTable()
	L.SetGlobal("req", req)

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

	dict := map[string]string{
		"id":     c.Locals("requestid").(string),
		"method": c.Method(),
		"url":    u.String(),
		"path":   "/" + c.Params("*"),
		"body":   string(c.Body()),
	}
	for k, v := range dict {
		req.RawSetString(k, lua.LString(v))
	}
	query := L.NewTable()
	q.VisitAll(func(key, value []byte) {
		query.RawSetString(string(key), lua.LString(value))
	})
	req.RawSetString("query", query)

	headers := L.NewTable()
	req.RawSetString("headers", headers)
	for k, v := range c.GetReqHeaders() {
		headers.RawSetString(k, lua.LString(v))
	}
}

func registerResponse(L *lua.LState, c *fiber.Ctx) {
	resp := L.NewTable()
	L.SetGlobal("resp", resp)

	headers := L.NewTable()
	resp.RawSetString("headers", headers)
	mtHeaders := L.NewTable()
	mtHeaders.RawSetString("__setindex", L.NewFunction(func(L *lua.LState) int {
		c.Set(L.CheckString(2), L.ToString(3))
		return 0
	}))
	L.SetMetatable(headers, mtHeaders)

	resp.RawSetString("type", L.NewFunction(func(L *lua.LState) int {
		if L.GetTop() > 1 {
			c.Type(L.CheckString(1), L.CheckString(2))
		} else {
			c.Type(L.CheckString(1))
		}
		return 0
	}))
	resp.RawSetString("send", L.NewFunction(func(L *lua.LState) int {
		status := 200
		body := ""
		if L.GetTop() > 1 {
			status = L.CheckInt(1)
			body = L.CheckString(2)
		} else {
			body = L.CheckString(1)
		}
		if err := c.Status(status).SendString(body); err != nil {
			L.RaiseError("http send failed: %v", err)
		}
		return 0
	}))
	resp.RawSetString("redir", L.NewFunction(func(L *lua.LState) int {
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
	}))
}

func registerSession(L *lua.LState, s *session.Session) {
	sess := L.NewTable()
	L.SetGlobal("sess", sess)
	mt := L.NewTable()
	L.SetMetatable(sess, mt)
	mt.RawSetString("__index", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(2)
		value := s.Get(key)
		L.Push(luar.New(L, value))
		return 1
	}))
	mt.RawSetString("__newindex", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(2)
		if L.Get(3) == lua.LNil {
			s.Delete(key)
			return 0
		}
		value := L.Get(3)
		goval := gluamapper.ToGoValue(value, gluamapper.Option{
			NameFunc: func(s string) string { return s },
		})
		s.Set(key, goval)
		return 0
	}))
	sess.RawSetString("keys", L.NewFunction(func(L *lua.LState) int {
		k := L.NewTable()
		for _, key := range s.Keys() {
			k.Append(lua.LString(key))
		}
		L.Push(k)
		return 1
	}))
	sess.RawSetString("save", L.NewFunction(func(L *lua.LState) int {
		t := L.ToNumber(1)
		if t != 0 {
			s.SetExpiry(time.Duration(t) * time.Hour)
		}
		if err := s.Save(); err != nil {
			L.RaiseError("session save failed: %v", err)
		}
		return 0
	}))
	sess.RawSetString("clear", L.NewFunction(func(L *lua.LState) int {
		if err := s.Destroy(); err != nil {
			L.RaiseError("session clear failed: %v", err)
		}
		return 0
	}))
}

func registerDB(L *lua.LState, c config.DB) {
	db, err := luasql.Open(L, c.Driver, c.Conn)
	if err != nil {
		L.RaiseError("db open failed: %v", err)
	}
	L.SetGlobal("db", db)
}

func hardStop(termCh chan os.Signal, stopCh chan any) {
	select {
	case <-termCh:
		// terminate
		os.Exit(1)
	case <-stopCh:
		return
	}
}
