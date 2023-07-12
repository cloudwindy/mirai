package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"mirai/pkg/config"
	"mirai/pkg/leapp"
	"mirai/pkg/lecli"
	"mirai/pkg/ledb"
	"mirai/pkg/luaengine"

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
	Version = "1.1"
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
)

func main() {
	// 设置退出信号
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(term)

	// 优雅退出
	done := make(chan any, 1)
	go hardStop(term, done)

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
	c, err := config.Parse(ctx.Path("proj"))
	if err != nil {
		return err
	}

	color.Blue(art.String("Mirai Project"))
	fmt.Println(" Mirai Server " + Version + " with " + luaengine.LuaVersion)
	fmt.Println(" Fiber " + fiber.Version)
	fmt.Println()

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
		Database: path.Join(c.Data, "fiber.db"),
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

	admin := api.Group("/admin")
	if c.Editing {
		admin.All("/files/*", filesHandler(c.Index))
		fmt.Print(" Editing: ")
		warn("Allowed")
		fmt.Println()
	}

	start := func(child *fiber.App) error {
		api.Mount("/", child).Name("app")

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
		return errors.Wrap(app.Listen(c.Listen), "http start")
	}

	engine := luaengine.New(c.Index, c.Env)
	capp := leapp.Config{
		Globals: []string{"db", "app"},
		Store:   store,
		Start:   start,
	}
	engine.Register("app", leapp.New(capp))
	engine.Register("db", ledb.New(c.DB))
	engine.Register("cli", lecli.New(colors))
	if err := engine.Run(); err != nil {
		return err
	}

	return nil
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

func filesHandler(base string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file := c.Params("*")
		if file == "" {
			names := make([]string, 0)
			err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
				if !d.IsDir() {
					path = strings.TrimPrefix(path, base+"/")
					names = append(names, path)
				}
				return err
			})
			if err != nil {
				return err
			}
			return c.JSON(names)
		}
		file = path.Join(base, file)
		switch c.Method() {
		case "GET":
			return c.SendFile(file)
		case "PUT":
			ensureDir(file)
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

// Create directories recursively
func ensureDir(fileName string) {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			panic(merr)
		}
	}
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

func LogTime(t time.Time) {
	fmt.Printf("%.02fms\n", float64(time.Since(t).Microseconds())/1000)
}
