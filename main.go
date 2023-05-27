package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/session"
	sbolt "github.com/gofiber/storage/bbolt"
	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
	"github.com/zs5460/art"
	bolt "go.etcd.io/bbolt"
)

const (
	Version       = "1.0"
	Listen        = ":3000"
	EnableEditing = true
)

//go:embed build/*
var BuildFS embed.FS

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

	app := fiber.New(fiber.Config{
		ServerHeader:          "Mirai Server",
		DisableStartupMessage: true,
	})
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

	db, err := bolt.Open("db/data.db", 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		panic(err)
	}

	storage := sbolt.New(sbolt.Config{
		Database: "db/fiber.db",
	})
	store := session.New(session.Config{
		Storage: storage,
	})

	api := app.Group("/api")
	api.Use(limiter.New(limiter.Config{
		Max:        200,
		Expiration: 10 * time.Minute,
		Storage:    storage,
	}))
	api.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))
	api.All("/v1/:scriptname?/*", luaHandler("./v1", db, store))
	admin := api.Group("/admin")
	if EnableEditing {
		fmt.Printf(" Editing: %s\n", color.GreenString("Enabled"))
		admin.All("/files/:path?", filesHandler("./v1"))
	}

	app.Use(etag.New())
	app.Use(cache.New(cache.Config{
		CacheHeader:  "Cache-Status",
		CacheControl: true,
		Expiration:   72 * time.Hour,
	}))
	app.Use(filesystem.New(filesystem.Config{
		Root:       http.FS(BuildFS),
		PathPrefix: "build",
	}))
	app.Get("*", func(c *fiber.Ctx) error {
		c.Path("/")
		return c.RestartRouting()
	})

	fmt.Println("\n Listening at", color.BlueString(Listen))
	if err := app.Listen(Listen); err != nil {
		panic(errors.Wrap(err, "无法启动 HTTP 服务器"))
	}
}

// Global LState pool
var (
	filesStat = make(map[string]fs.FileInfo)
	luaCache  = make(map[string]*lua.FunctionProto)
	luaPool   = &lStatePool{
		saved: make([]*lua.LState, 0, 4),
	}
)

func luaHandler(base string, db *bolt.DB, store *session.Store) fiber.Handler {
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
			proto, err := CompileLua(file)
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

		registerRequest(L, c)
		registerResponse(L, c)
		registerSession(L, s)
		registerKVStore(L, db)

		if err := DoCompiledFile(L, luaCache[file]); err != nil {
			if lerr := err.(*lua.ApiError); lerr != nil {
				c.Locals("stacktrace", lerr.StackTrace)
				return &apiError{msg: lerr.Object.String()}
			}
			return err
		}

		return nil
	}
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

func registerRequest(L *lua.LState, c *fiber.Ctx) {
	req := L.NewTable()
	L.SetGlobal("req", req)

	L.SetField(req, "id", lua.LString(c.Locals("requestid").(string)))
	L.SetField(req, "method", lua.LString(c.Method()))
	L.SetField(req, "url", lua.LString(c.OriginalURL()))
	L.SetField(req, "protocol", lua.LString(c.Protocol()))
	L.SetField(req, "host", lua.LString(c.Hostname()))
	L.SetField(req, "port", lua.LString(c.Port()))
	L.SetField(req, "path", lua.LString(c.Path()))
	L.SetField(req, "subpath", lua.LString(c.Params("*")))
	L.SetField(req, "body", lua.LString(c.Body()))

	query := L.NewTable()
	L.SetField(req, "query", query)
	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		L.SetField(query, string(key), lua.LString(value))
	})

	headers := L.NewTable()
	L.SetField(req, "headers", headers)
	for k, v := range c.GetReqHeaders() {
		L.SetField(headers, k, lua.LString(v))
	}
}

func registerResponse(L *lua.LState, c *fiber.Ctx) {
	resp := L.NewTable()
	L.SetGlobal("resp", resp)

	headers := L.NewTable()
	L.SetField(resp, "headers", headers)
	mtHeaders := L.NewTable()
	L.SetField(mtHeaders, "__setindex", L.NewFunction(func(L *lua.LState) int {
		c.Set(L.CheckString(2), L.ToString(3))
		return 0
	}))
	L.SetMetatable(headers, mtHeaders)

	L.SetField(resp, "type", L.NewFunction(func(L *lua.LState) int {
		if L.GetTop() > 1 {
			c.Type(L.CheckString(1), L.CheckString(2))
		} else {
			c.Type(L.CheckString(1))
		}
		return 0
	}))
	L.SetField(resp, "send", L.NewFunction(func(L *lua.LState) int {
		status := 200
		body := ""
		if L.GetTop() > 1 {
			status = L.CheckInt(1)
			body = L.CheckString(2)
		} else {
			body = L.CheckString(1)
		}
		if err := c.Status(status).SendString(body); err != nil {
			L.Error(lua.LString(err.Error()), 1)
		}
		return 0
	}))
	L.SetField(resp, "redir", L.NewFunction(func(L *lua.LState) int {
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

	L.SetField(mt, "__index", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(2)
		value, ok := s.Get(key).(string)
		if !ok {
			L.Push(lua.LNil)
		}
		L.Push(lua.LString(value))
		return 1
	}))
	L.SetField(mt, "__newindex", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(2)
		value := L.ToString(3)
		s.Set(key, value)
		return 0
	}))
	L.SetField(sess, "keys", L.NewFunction(func(L *lua.LState) int {
		k := L.NewTable()
		for _, key := range s.Keys() {
			k.Append(lua.LString(key))
		}
		L.Push(k)
		return 1
	}))
	L.SetField(sess, "save", L.NewFunction(func(L *lua.LState) int {
		if err := s.Save(); err != nil {
			L.Error(lua.LString(err.Error()), 1)
		}
		return 0
	}))
	L.SetField(sess, "clear", L.NewFunction(func(L *lua.LState) int {
		if err := s.Destroy(); err != nil {
			L.Error(lua.LString(err.Error()), 1)
		}
		return 0
	}))
}

type Bucket struct {
	DB   *bolt.DB
	Name string
}

func registerKVStore(L *lua.LState, db *bolt.DB) {
	mt := L.NewTable()
	L.SetGlobal("kv", mt)
	L.SetField(mt, "new", L.NewFunction(func(L *lua.LState) int {
		b := new(Bucket)
		b.DB = db
		b.Name = L.CheckString(1)
		ud := L.NewUserData()
		ud.Value = b
		L.SetMetatable(ud, mt)
		L.Push(ud)
		return 1
	}))
	L.SetField(mt, "__index", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(2)
		if fn, ok := lkvExports[key]; ok {
			L.Push(L.NewFunction(fn))
			return 1
		}
		return lkvGet(L)
	}))
	L.SetField(mt, "__newindex", L.NewFunction(lkvPut))
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
