package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pkg/errors"
	libs "github.com/vadv/gopher-lua-libs"
	lua "github.com/yuin/gopher-lua"
	"github.com/zs5460/art"
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

	app := fiber.New(fiber.Config{
		ServerHeader:          "Mirai Server",
		DisableStartupMessage: true,
	})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(favicon.New())
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

	api := app.Group("/api")
	api.Get("/", func(c *fiber.Ctx) error {
		L := lua.NewState()
		defer L.Close()
		libs.Preload(L)
		registerRequest(c, L)
		registerResponse(c, L)
		if err := L.DoFile("lua/index.lua"); err != nil {
			return err
		}
		return nil
	})

	app.Static("/", "build", fiber.Static{
		Compress:  true,
		ByteRange: true,
	})

	app.All("*", func(c *fiber.Ctx) error {
		return c.Status(404).SendFile("build/index.html", true)
	})

	fmt.Println(art.String("Mirai Server"))

	if err := app.Listen(":3000"); err != nil {
		panic(errors.Wrap(err, "无法启动 HTTP 服务器"))
	}
}

func registerRequest(c *fiber.Ctx, L *lua.LState) {
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

func registerResponse(c *fiber.Ctx, L *lua.LState) {
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

func hardStop(termCh chan os.Signal, stopCh chan any) {
	select {
	case <-termCh:
		// terminate
		os.Exit(1)
	case <-stopCh:
		return
	}
}
