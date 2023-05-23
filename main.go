package main

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"mirai/modules/libs"

	"github.com/boltdb/bolt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
	"github.com/zs5460/art"
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

	fmt.Println(art.String("Mirai Project"))
	fmt.Println(" Mirai Server " + Version + " with " + lua.LuaVersion)
	fmt.Println(" Fiber " + fiber.Version)

	app := fiber.New(fiber.Config{
		ServerHeader:          "Mirai Server",
		DisableStartupMessage: true,
	})
	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(favicon.New())
	app.Use(compress.New())
	skipApi := func(c *fiber.Ctx) bool {
		return strings.HasPrefix(c.Path(), "/api")
	}
	app.Use(etag.New(etag.Config{
		Next: skipApi,
	}))
	app.Use(cache.New(cache.Config{
		Next:         skipApi,
		CacheControl: true,
		CacheHeader:  "Cache-Status",
		Expiration:   24 * time.Hour,
	}))

	db, err := bolt.Open("data.db", 0o600, bolt.DefaultOptions)
	if err != nil {
		panic(err)
	}

	api := app.Group("/api")
	api.All("/v1/:scriptname?/*", luaHandler("./v1", db))
	admin := api.Group("/admin")
	if EnableEditing {
		fmt.Println(" Editing: Enabled")
		admin.All("/files/:subpath", filesHandler("./v1"))
	}

	app.Use(filesystem.New(filesystem.Config{
		Root:       http.FS(BuildFS),
		PathPrefix: "build",
	}))

	app.Get("*", func(c *fiber.Ctx) error {
		f, err := BuildFS.ReadFile("build/index.html")
		if err != nil {
			return err
		}
		return c.Status(200).Type("html").Send(f)
	})

	fmt.Println()
	fmt.Println(" Listening at " + Listen)
	fmt.Println()
	if err := app.Listen(Listen); err != nil {
		panic(errors.Wrap(err, "无法启动 HTTP 服务器"))
	}
}

type lStatePool struct {
	m     sync.Mutex
	saved []*lua.LState
}

func (pl *lStatePool) Get() *lua.LState {
	pl.m.Lock()
	defer pl.m.Unlock()
	n := len(pl.saved)
	if n == 0 {
		return pl.New()
	}
	x := pl.saved[n-1]
	pl.saved = pl.saved[0 : n-1]
	return x
}

func (pl *lStatePool) New() *lua.LState {
	L := lua.NewState()
	// setting the L up here.
	// load scripts, set global variables, share channels, etc...
	libs.PreloadAll(L)
	return L
}

func (pl *lStatePool) Put(L *lua.LState) {
	pl.m.Lock()
	defer pl.m.Unlock()
	pl.saved = append(pl.saved, L)
}

func (pl *lStatePool) Shutdown() {
	for _, L := range pl.saved {
		L.Close()
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

func luaHandler(base string, db *bolt.DB) fiber.Handler {
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

		registerRequest(L, c)
		registerResponse(L, c)
		registerKVStore(L, db)
		DoCompiledFile(L, luaCache[file])

		if err := L.DoFile(file); err != nil {
			return err
		}
		return nil
	}
}

// CompileLua reads the passed lua file from disk and compiles it.
func CompileLua(filePath string) (*lua.FunctionProto, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	chunk, err := parse.Parse(reader, filePath)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, err
	}
	return proto, nil
}

func DoCompiledFile(L *lua.LState, proto *lua.FunctionProto) error {
	lfunc := L.NewFunctionFromProto(proto)
	L.Push(lfunc)
	return L.PCall(0, lua.MultRet, nil)
}

func filesHandler(base string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file := c.Params("subpath")
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

	url := lua.LString(c.OriginalURL())
	L.SetField(req, "url", url)

	path := lua.LString(c.Path())
	L.SetField(req, "path", path)

	subpath := lua.LString(c.Params("*"))
	L.SetField(req, "subpath", subpath)

	query := L.NewTable()
	L.SetField(req, "query", query)
	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		L.SetField(query, string(key), lua.LString(key))
	})

	method := lua.LString(c.Method())
	L.SetField(req, "method", method)

	headers := L.NewTable()
	L.SetField(req, "headers", headers)
	for k, v := range c.GetReqHeaders() {
		L.SetField(headers, k, lua.LString(v))
	}

	body := lua.LString(c.Body())
	L.SetField(req, "body", body)

	get := L.NewFunction(func(l *lua.LState) int {
		value := c.Get(l.CheckString(1))
		l.Push(lua.LString(value))
		return 1
	})
	L.SetField(req, "get", get)
}

func registerResponse(L *lua.LState, c *fiber.Ctx) {
	resp := L.NewTable()
	L.SetGlobal("resp", resp)

	set := L.NewFunction(func(l *lua.LState) int {
		c.Set(l.CheckString(1), l.CheckString(2))
		return 0
	})
	L.SetField(resp, "set", set)

	_type := L.NewFunction(func(l *lua.LState) int {
		if l.GetTop() > 1 {
			c.Type(l.CheckString(1), l.CheckString(2))
		} else {
			c.Type(l.CheckString(1))
		}
		return 0
	})
	L.SetField(resp, "type", _type)

	send := L.NewFunction(func(l *lua.LState) int {
		status := 200
		body := ""
		if l.GetTop() > 1 {
			status = l.CheckInt(1)
			body = l.CheckString(2)
		} else {
			body = l.CheckString(1)
		}
		if err := c.Status(status).SendString(body); err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		return 0
	})
	L.SetField(resp, "send", send)

	redir := L.NewFunction(func(l *lua.LState) int {
		status := 302
		location := ""
		if l.GetTop() > 1 {
			status = l.CheckInt(1)
			location = l.CheckString(2)
		} else {
			location = l.CheckString(1)
		}
		c.Status(status).Location(location)
		return 0
	})
	L.SetField(resp, "redir", redir)
}

var delimiter = ":"

func registerKVStore(L *lua.LState, db *bolt.DB) {
	kv := L.NewTable()
	L.SetGlobal("kv", kv)
	get := L.NewFunction(func(l *lua.LState) int {
		bucket := l.CheckString(1)
		key := l.CheckString(2)

		tx, err := db.Begin(false)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		defer tx.Rollback()

		parts := strings.Split(bucket, delimiter)
		b := tx.Bucket([]byte(parts[0]))
		if b == nil {
			L.Push(lua.LNil)
			return 1
		}
		for _, part := range parts[1:] {
			b = b.Bucket([]byte(part))
			if b == nil {
				L.Push(lua.LNil)
				return 1
			}
		}
		L.Push(lua.LString(b.Get([]byte(key))))
		return 1
	})
	L.SetField(kv, "get", get)
	set := L.NewFunction(func(l *lua.LState) int {
		bucket := l.CheckString(1)
		key := l.CheckString(2)
		value := l.CheckString(3)

		tx, err := db.Begin(true)
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		defer tx.Rollback()

		parts := strings.Split(bucket, delimiter)
		b, err := tx.CreateBucketIfNotExists([]byte(parts[0]))
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		for _, part := range parts[1:] {
			b, err = b.CreateBucketIfNotExists([]byte(part))
			if err != nil {
				L.Push(lua.LString(err.Error()))
				return 1
			}
		}
		err = b.Put([]byte(key), []byte(value))
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		if err = tx.Commit(); err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		return 0
	})
	L.SetField(kv, "set", set)
	drop := L.NewFunction(func(l *lua.LState) int {
		bucket := l.CheckString(1)

		tx, err := db.Begin(true)
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		defer tx.Rollback()

		if err = tx.DeleteBucket([]byte(bucket)); err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		if err = tx.Commit(); err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		return 0
	})
	L.SetField(kv, "drop", drop)
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
