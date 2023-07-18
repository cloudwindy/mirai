package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"mirai/pkg/admin"
	"mirai/pkg/config"
	"mirai/pkg/leapp"
	"mirai/pkg/lecli"
	"mirai/pkg/ledb"
	"mirai/pkg/lue"
	"mirai/pkg/timer"

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
	"github.com/urfave/cli/v2"
	"github.com/zs5460/art"
)

// Package info
const (
	Version = "1.2"
)

// Color helper functions
var (
	colors = map[string]lecli.Print{
		"succ": succ,
		"info": info,
		"warn": warn,
	}
	succ = color.New(color.FgGreen).PrintFunc()
	info = color.New(color.FgBlue).PrintFunc()
	warn = color.New(color.FgHiYellow).PrintFunc()
	//go:embed ilua.lua
	ilua string
)

func main() {
	app := &cli.App{
		Name:                   "mirai",
		Usage:                  "Server for the Mirai Project",
		Version:                Version,
		DefaultCommand:         "start",
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "proj",
				Aliases: []string{"p"},
				Usage:   "project directory",
				Value:   ".",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "start the server (default)",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "edit",
						Aliases: []string{"e"},
						Usage:   "allow editing",
					},
				},
				Action: start,
			},
			{
				Name:  "run",
				Usage: "run command",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		color.Red("%v", err)
	}
}

func start(ctx *cli.Context) error {
	// 设置退出信号
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(term)

	// 优雅退出
	done := make(chan any, 1)
	go hardStop(term, done)

	c, err := config.Parse(ctx.Path("proj"))
	if err != nil {
		return err
	}

	color.Blue(art.String("Mirai Project"))
	fmt.Println(" Mirai Server " + Version + " with " + lue.LuaVersion)
	fmt.Println(" Fiber " + fiber.Version)
	fmt.Println()

	app := fiber.New(fiber.Config{
		ServerHeader:          "Mirai Server",
		DisableStartupMessage: true,
	})
	app.Use(timer.Print("total", "Total Time"))
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
		Database: path.Join(c.Data, "fiber.db"),
	})
	store := session.New(session.Config{
		Storage: storage,
	})

	apigrp := app.Group("/api")
	if l := c.Limiter; l.Enabled {
		apigrp.Use(limiter.New(limiter.Config{
			Max:        l.Max,
			Expiration: time.Duration(l.Dur) * time.Second,
		}))
	}
	apigrp.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))
	apigrp.Use(timer.Print("exec", "Script Execution"))

	admingrp := apigrp.Group("/admin")
	if c.Editing {
		admingrp.All("/files/*", admin.Files(c.Index))
		fmt.Print(" Editing: ")
		warn("Allowed")
		fmt.Println()
	}

	start := func(child *fiber.App) {
		apigrp.Mount("/", child).Name("app")

		app.Use(etag.New())
		app.Use(cache.New(cache.Config{
			Storage:      storage,
			CacheHeader:  "Cache-Status",
			CacheControl: false,
			Expiration:   72 * time.Hour,
		}))
		if c.Root != "" {
			app.Static("/", c.Root, fiber.Static{
				ByteRange: true,
			})
			app.Get("*", func(c *fiber.Ctx) error {
				c.Path("/")
				return c.RestartRouting()
			})
		}

		fmt.Println()
		fmt.Print(" Listening at ")
		info(c.Listen)
		fmt.Println()
		fmt.Println()
		go func() {
			if err := app.Listen(c.Listen); err != nil {
				panic(errors.Wrap(err, "http start"))
			}
		}()
	}

	engine := lue.New(c.Index, c.Env)
	capp := leapp.Config{
		Store: store,
		Start: start,
	}
	engine.Register("app", leapp.New(capp)).
		Register("db", ledb.New(c.DB)).
		Register("cli", lecli.New(colors)).
		Run()

	done <- nil
	engine.Eval(ilua).
		Eval(`
		local params = {
			prompt = '> ',
			prompt2 = '  ',
			disable_startup_message = true
		}
		local ilua = Ilua:new(params)
		ilua:start()
		ilua:run()
		`)

	return nil
}

func hardStop(term chan os.Signal, stop chan any) {
	select {
	case <-term:
		// terminate
		os.Exit(1)
	case <-stop:
		return
	}
}

func LogTime(t time.Time) {
	fmt.Printf("%.02fms\n", float64(time.Since(t).Microseconds())/1000)
}
