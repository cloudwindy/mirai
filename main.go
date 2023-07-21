package main

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	"mirai/pkg/admin"
	"mirai/pkg/config"
	"mirai/pkg/daemon"
	"mirai/pkg/dir"
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
	"github.com/inancgumus/screen"
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
		"fail": fail,
	}
	succ = color.New(color.FgGreen).PrintFunc()
	info = color.New(color.FgBlue).PrintFunc()
	warn = color.New(color.FgHiYellow).PrintFunc()
	fail = color.New(color.FgHiRed).PrintFunc()
	//go:embed ilua.lua
	ilua string
)

func main() {
	time.Sleep(100 * time.Millisecond)
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
					&cli.PathFlag{
						Name:  "pidfile",
						Usage: "file which the child's pid is stored in",
						Value: path.Join(os.TempDir(), "mirai.pid"),
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
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	c, err := config.Parse(ctx.Path("proj"))
	if err != nil {
		return err
	}

	if daemon.IsChild() {
		return worker(ctx, c)
	}

	pidpath := ctx.Path("pidfile")
	err = daemon.WritePid(pidpath)
	if err != nil {
		return err
	}

	ln, err := daemon.Forked(c.Listen)
	if err != nil {
		return err
	}

	_, err = daemon.Fork(wd, ln)
	if err != nil {
		return err
	}
	exit := make(chan bool)
	handler := func(sig os.Signal) {
		switch sig {
		case syscall.SIGHUP:
			_, err = daemon.Fork(wd, ln)
			if err != nil {
				fail(err)
			}
		case syscall.SIGTERM:
			exit <- true
		}
	}
	sigln := daemon.Listen(handler, syscall.SIGHUP, syscall.SIGTERM)
	<-exit
	sigln.Close()

	if err := os.Remove(pidpath); err != nil {
		return err
	}
	return nil
}

func worker(cliCtx *cli.Context, c config.Config) error {
	sig := daemon.Listen(daemon.ExitHandler, os.Interrupt, syscall.SIGTERM)

	pid, err := daemon.ReadPid(cliCtx.Path("pidfile"))
	if err != nil {
		return err
	}

	ln, err := daemon.Forked(c.Listen)
	if err != nil {
		return err
	}

	screen.Clear()
	screen.MoveTopLeft()

	color.Blue(art.String("Mirai Project"))
	fmt.Println(" Mirai Server " + Version + " with " + lue.LuaVersion)
	fmt.Println(" Fiber " + fiber.Version)
	fmt.Println()

	app := fiber.
		New(fiber.Config{
			ServerHeader:          "Mirai Server",
			DisableStartupMessage: true,
		})
	app.
		Use(timer.Print("total", "Total Time")).
		Use(favicon.New()).
		Use(requestid.New()).
		Use(logger.New(logger.Config{
			Done: func(c *fiber.Ctx, logString []byte) {
				if tb := c.Locals("stacktrace"); tb != nil {
					color.Red("%s", tb)
				}
			},
		})).
		Use(cors.New()).
		Use(compress.New()).
		Use(pprof.New())

	apigrp := app.Group("/api")
	if l := c.Limiter; l.Enabled {
		apigrp.Use(limiter.New(limiter.Config{
			Max:        l.Max,
			Expiration: time.Duration(l.Dur) * time.Second,
		}))
	}
	apigrp.
		Use(recover.New(recover.Config{
			EnableStackTrace: true,
		})).
		Use(timer.Print("exec", "Script Execution"))

	admingrp := apigrp.Group("/admin")
	if c.Editing {
		admingrp.All("/files/*", admin.Files(c.Index))
		fmt.Print(" Editing: ")
		warn("Allowed")
		fmt.Println()
	}

	storage := sbolt.New(sbolt.Config{
		Database: path.Join(c.Data, "fiber.db"),
	})
	store := session.New(session.Config{
		Storage: storage,
	})

	if c.Root != "" {
		ok, err := dir.Is(c.Root)
		if !ok {
			return errors.New("Root directory does not exist")
		}
		if err != nil {
			return err
		}
		const localSkip = "__main_skip"
		next := func(c *fiber.Ctx) bool {
			return c.Locals(localSkip).(bool)
		}
		app.
			Use(func(c *fiber.Ctx) error {
				c.Locals(localSkip, strings.HasPrefix(c.Path(), "/api"))
				return c.Next()
			}).
			Use(etag.New(etag.Config{
				Next: next,
			})).
			Use(cache.New(cache.Config{
				Next:         next,
				Storage:      storage,
				CacheHeader:  "Cache-Status",
				CacheControl: false,
				Expiration:   72 * time.Hour,
			})).
			Static("/", c.Root, fiber.Static{
				Next:      next,
				ByteRange: true,
			}).
			Get("*", func(c *fiber.Ctx) error {
				if next(c) {
					return c.Next()
				}
				c.Path("/")
				return c.RestartRouting()
			})
	}

	engine := lue.New(c.Index, c.Env)

	start := func() error {
		fmt.Print("\n Listening at ")
		info(c.Listen)
		fmt.Print("\n\n")
		go func() {
			if err := app.Listener(ln); err != nil {
				panic(errors.Wrap(err, "http start"))
			}
		}()
		return nil
	}

	reload := func() error {
		if runtime.GOOS == "windows" {
			return errors.New("not supported on windows")
		}
		fmt.Print("\n Reloading...")

		if err := syscall.Kill(pid, syscall.SIGHUP); err != nil {
			return err
		}

		os.Exit(0)
		return nil
	}

	stop := func(timeout time.Duration) error {
		fmt.Print("\n Shutting down...")
		defer fmt.Print("\n\n")

		sig := daemon.Listen(daemon.ExitHandler, os.Interrupt, syscall.SIGTERM)
		defer sig.Close()

		var err error
		if timeout != 0 {
			err = app.ShutdownWithTimeout(timeout)
		} else {
			err = app.Shutdown()
		}
		if err != nil {
			return err
		}

		return nil
	}

	capp := leapp.Config{
		App:    app.Group("/api"),
		Store:  store,
		Start:  start,
		Reload: reload,
		Stop:   stop,
	}
	engine.
		Register("app", leapp.New(capp)).
		Register("db", ledb.New(c.DB)).
		Register("cli", lecli.New(colors)).
		Run().
		Eval(ilua)

	if err := engine.Err(); err != nil {
		return err
	}

	exit := make(chan bool)

	go func() {
		engine.Eval(`
			local params = {
				prompt = '> ',
				prompt2 = '  ',
				disable_startup_message = true
			}
			local ilua = Ilua:new(params)
			ilua:start()
			ilua:run()
		`)

		if err := engine.Err(); err != nil {
			fail(err)
		}

		exit <- true
	}()

	sig.Close()
	<-exit
	syscall.Kill(pid, syscall.SIGTERM)

	return nil
}
