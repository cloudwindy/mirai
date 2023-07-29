package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/cloudwindy/mirai/pkg/admin"
	"github.com/cloudwindy/mirai/pkg/config"
	"github.com/cloudwindy/mirai/pkg/daemon"
	"github.com/cloudwindy/mirai/pkg/dir"
	"github.com/cloudwindy/mirai/pkg/leapp"
	"github.com/cloudwindy/mirai/pkg/lecli"
	"github.com/cloudwindy/mirai/pkg/ledb"
	"github.com/cloudwindy/mirai/pkg/lue"
	"github.com/cloudwindy/mirai/pkg/timer"
	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/session"
	sbolt "github.com/gofiber/storage/bbolt"
	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// Package info
const (
	Version    = "1.2"
	ServerName = "mirai/" + Version
)

// Color helper functions
var (
	colors = map[string]lecli.Printf{
		"print": print,
		"succ":  succ,
		"info":  info,
		"warn":  warn,
		"fail":  fail,
	}
	print = color.New(color.Reset).PrintfFunc()
	succ  = color.New(color.FgGreen).PrintfFunc()
	info  = color.New(color.FgBlue).PrintfFunc()
	warn  = color.New(color.FgYellow).PrintfFunc()
	fail  = color.New(color.FgRed).PrintfFunc()
)

//go:embed ilua.lua
var ilua string

var (
	sqlver, _, _ = sqlite3.Version()
	globalEnv    = map[string]any{
		"VERSION":        Version,
		"GO_VERSION":     strings.TrimPrefix(runtime.Version(), "go"),
		"FIBER_VERSION":  fiber.Version,
		"SQLITE_VERSION": sqlver,
	}
)

var DefaultPidFile = "mirai.pid"

func main() {
	time.Sleep(100 * time.Millisecond)
	app := cli.NewApp()
	app.Usage = "Server for the Mirai Project"
	app.Version = Version
	app.DefaultCommand = "start"
	app.UseShortOptionHandling = true
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		&cli.PathFlag{
			Name:    "proj",
			Aliases: []string{"p"},
			Usage:   "set project path",
			Value:   ".",
		},
		&cli.BoolFlag{
			Name:    "interactive",
			Aliases: []string{"i"},
			Usage:   "enter interactive mode after executing",
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:  "start",
			Usage: "Start the server (default)",
			Description: "Start command finds project.lua in the specified project path. \n" +
				"Then, it starts the server based on the configuration. \n" +
				"If it is not found, it will set up a temporary server and database and enter interactive mode.",
			ArgsUsage:       "arguments are passed to Lua scripts without parsing",
			SkipFlagParsing: true,
			Action:          start,
		},
		{
			Name:            "run",
			Usage:           "Run command",
			ArgsUsage:       "arguments are passed to Lua scripts without parsing",
			SkipFlagParsing: true,
			Action:          run,
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
		if errors.Is(err, os.ErrNotExist) {
			startInteractive(ctx)
			return nil
		}
		return err
	}
	for k, v := range c.Env {
		globalEnv[k] = v
	}

	if c.Pid == "" {
		c.Pid = path.Join(os.TempDir(), DefaultPidFile)
	}
	if daemon.IsChild() || runtime.GOOS == "windows" {
		return worker(ctx, c)
	}

	if err = daemon.WritePid(c.Pid); err != nil {
		return err
	}
	defer os.Remove(c.Pid)

	ln, err := daemon.Forked(c.Listen)
	if err != nil {
		return err
	}

	var proc, newproc *os.Process
	proc, err = daemon.Fork(wd, ln)
	if err != nil {
		return err
	}
	if proc == nil {
		return worker(ctx, c)
	}
	handler := func(sig os.Signal) {
		switch sig {
		case syscall.SIGHUP:
			if err := proc.Kill(); err != nil {
				fail("%v\n", err)
			}
			if newproc, err = daemon.Fork(wd, ln); err != nil {
				fail("%v\n", err)
			}
		case syscall.SIGTERM:
			if err = proc.Kill(); err != nil {
				fail("%v\n", err)
			}
		}
	}
	sigln := daemon.Listen(handler, syscall.SIGHUP, syscall.SIGTERM)
	defer sigln.Close()

	if _, err := proc.Wait(); err != nil {
		return err
	}
	for proc != newproc {
		proc = newproc
		_, err := proc.Wait()
		if err != nil {
			return err
		}
	}

	return nil
}

func worker(ctx *cli.Context, c config.Config) error {
	sigln := daemon.Listen(daemon.ExitHandler, os.Interrupt)

	ln, err := daemon.Forked(c.Listen)
	if err != nil {
		return err
	}

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
		Use(pprof.New()).
		Use(func(c *fiber.Ctx) error {
			// set before next to allow modifying
			c.Set("Server", ServerName)
			return c.Next()
		})

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
		fmt.Print("Editing: ")
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
			return errors.New("root directory does not exist")
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
			Use(cache.New(cache.Config{
				Next:         next,
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

	start := func(_ string) error {
		go func() {
			if err := app.Listener(ln); err != nil {
				panic(errors.Wrap(err, "http start"))
			}
		}()
		return nil
	}

	var reload func() error
	if daemon.IsChild() {
		pid, err := daemon.ReadPid(c.Pid)
		if err != nil {
			return err
		}
		reload = func() error {
			fmt.Println("reloading...")

			if err := daemon.Kill(pid, syscall.SIGHUP); err != nil {
				return err
			}

			os.Exit(0)
			return nil
		}
	}

	stop := func(timeout time.Duration) error {
		fmt.Print("\nshutting down...")
		defer fmt.Println()

		sig := daemon.Listen(daemon.ExitHandler, os.Interrupt)
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

	G := lue.New(globalEnv).
		Register("app", leapp.New(capp)).
		Register("db", ledb.New(c.DB)).
		Register("cli", lecli.New(ctx.Args().Slice(), colors)).
		Run(c.Index)

	if err := G.Err(); err != nil {
		return err
	}

	sigln.Close()
	if ctx.Bool("interactive") {
		interactive(G)
	} else {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGQUIT)
		<-c
		G.Eval(`app:stop(10)`)
	}

	return nil
}

func startInteractive(ctx *cli.Context) {
	fmt.Printf("Mirai Server %s with %s\n", Version, lue.LuaVersion)
	app := fiber.New()
	store := session.New()
	capp := leapp.Config{
		App:   fiber.New(),
		Store: store,
		Start: func(listen string) error {
			go func() {
				if err := app.Listen(listen); err != nil {
					fail("%v\n", err)
				}
			}()
			return nil
		},
		Stop: func(timeout time.Duration) error {
			if timeout != 0 {
				return app.ShutdownWithTimeout(timeout)
			} else {
				return app.Shutdown()
			}
		},
	}
	db := config.DB{
		Driver: "sqlite3",
		Conn:   ":memory:",
	}
	G := lue.New(globalEnv).
		Register("app", leapp.New(capp)).
		Register("db", ledb.New(db)).
		Register("cli", lecli.New(ctx.Args().Slice(), colors))
	if err := G.Err(); err != nil {
		fail("%v\n", err)
	}
	interactive(G)
}

func run(ctx *cli.Context) error {
	args := ctx.Args()
	if args.Len() < 1 {
		return errors.New("command not specified")
	}
	c, err := config.Parse(ctx.Path("proj"))
	if err != nil {
		return err
	}
	for k, v := range c.Env {
		globalEnv[k] = v
	}
	cmd, ok := c.Commands[args.First()]
	if !ok {
		return errors.New("command not found")
	}
	G := lue.New(globalEnv).
		Register("db", ledb.New(c.DB)).
		Register("cli", lecli.New(args.Tail(), colors)).
		Run(cmd)
	if err := G.Err(); err != nil {
		fail("%v\n", err)
	}
	return nil
}

func interactive(G *lue.Engine) {
	G.
		Eval(ilua).
		Eval(`
			local params = {
				prompt = '> ',
				prompt2 = '  ',
				disable_startup_message = true
			}
			local ilua = Ilua:new(params)
			ilua:start()
			ilua:run()
		`).
		Eval(`app:stop(10)`)
	if err := G.Err(); err != nil {
		fail("%v\n", err)
	}
}
