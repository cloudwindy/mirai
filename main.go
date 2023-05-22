package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pkg/errors"
	libJson "github.com/vadv/gopher-lua-libs/json"
	lua "github.com/yuin/gopher-lua"
	"github.com/zs5460/art"
)

const (
	Version       = "1.0"
	Listen        = ":3000"
	EnableEditing = true
)

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
	app.Use(recover.New())
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
	api.All("/v1/:subpath", luaHandler("./v1", db))
	admin := api.Group("/admin")
	if EnableEditing {
		fmt.Println(" Editing: Enabled")
		files := admin.Group("/files")
		files.Get("/:subpath", downloadHandler("./v1"))
		files.Put("/:subpath", uploadHandler("./v1"))
	}

	app.Static("/", "build")

	app.All("*", func(c *fiber.Ctx) error {
		return c.Status(200).SendFile("build/index.html", true)
	})

	fmt.Println()
	fmt.Println(" Listening at " + Listen)
	fmt.Println()
	if err := app.Listen(Listen); err != nil {
		panic(errors.Wrap(err, "无法启动 HTTP 服务器"))
	}
}

func luaHandler(base string, db *bolt.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file := c.Params("subpath")
		file = path.Join(base, file)
		info, err := os.Stat(file)
		if os.IsNotExist(err) {
			file += ".lua"
		}
		if info.IsDir() {
			file = path.Join(file, "index.lua")
		}
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fiber.ErrNotFound
		}

		L := lua.NewState()
		defer L.Close()
		libJson.Preload(L)
		registerRequest(L, c)
		registerResponse(L, c)
		registerDatabase(L, db)

		if err := L.DoFile(file); err != nil {
			return err
		}
		return nil
	}
}

func downloadHandler(base string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file := c.Params("subpath")
		file = path.Join(base, file)
		return c.SendFile(file)
	}
}

func uploadHandler(base string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file := c.Params("subpath")
		file = path.Join(base, file)
		if err := os.WriteFile(file, c.Body(), 0o644); err != nil {
			return err
		}
		return c.SendString("ok")
	}
}

func registerRequest(L *lua.LState, c *fiber.Ctx) {
	req := L.NewTable()
	L.SetGlobal("req", req)
	url := lua.LString(c.OriginalURL())
	L.SetField(req, "url", url)
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
	send := L.NewFunction(func(l *lua.LState) int {
		if err := c.SendString(l.CheckString(1)); err != nil {
			L.Error(lua.LString(err.Error()), 0)
		}
		return 0
	})
	L.SetField(resp, "send", send)
	sendJson := L.NewFunction(func(l *lua.LState) int {
		if err := c.JSON(l.CheckTable(1)); err != nil {
			L.Error(lua.LString(err.Error()), 0)
		}
		return 0
	})
	L.SetField(resp, "send_json", sendJson)
}

func registerDatabase(L *lua.LState, db *bolt.DB) {
	luaDb := L.NewTable()
	L.SetGlobal("db", luaDb)
	get := L.NewFunction(func(l *lua.LState) int {
		err := db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(l.CheckString(1)))
			if err != nil {
				return err
			}
			value := b.Get([]byte(l.CheckString(2)))
			L.Push(lua.LString(value))
			return nil
		})
		if err != nil {
			L.Error(lua.LString(err.Error()), 0)
		}
		return 1
	})
	L.SetField(luaDb, "get", get)
	set := L.NewFunction(func(l *lua.LState) int {
		err := db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(l.CheckString(1)))
			if err != nil {
				return err
			}
			return b.Put([]byte(l.CheckString(2)), []byte(l.CheckString(3)))
		})
		if err != nil {
			L.Error(lua.LString(err.Error()), 0)
		}
		return 0
	})
	L.SetField(luaDb, "set", set)
	drop := L.NewFunction(func(l *lua.LState) int {
		err := db.Update(func(tx *bolt.Tx) error {
			return tx.DeleteBucket([]byte(l.CheckString(1)))
		})
		if err != nil {
			L.Error(lua.LString(err.Error()), 0)
		}
		return 0
	})
	L.SetField(luaDb, "drop", drop)
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
