package main

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"regexp"
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
	"github.com/urfave/cli/v3"
)

// Package info
var (
	// initialized by Makefile
	version = "dev"
	build   = ""
)

var (
	version_match = regexp.MustCompile(`v([^-]*)`)
	version_clean = version_match.FindStringSubmatch(version)
	servername    = "mirai/" + version_clean[1]
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
		"OS":             runtime.GOOS,
		"ARCH":           runtime.GOARCH,
		"VERSION":        version,
		"GO_VERSION":     strings.TrimPrefix(runtime.Version(), "go"),
		"FIBER_VERSION":  fiber.Version,
		"SQLITE_VERSION": sqlver,
	}
)

var DefaultPidFile = "mirai.pid"

func main() {
	app := new(cli.Command)
	app.Usage = "Server for the Mirai Project"
	app.Version = fmt.Sprintf("%s %s", version, build)
	app.DefaultCommand = "start"
	app.UseShortOptionHandling = true
	app.EnableShellCompletion = true
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "proj",
			Aliases: []string{"p"},
			Usage:   "set project path",
			Value:   ".",
			Sources: cli.Files("."),
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
			Description: "Start command finds project.lua in the specified project path.\n" +
				"Then, it starts the server based on the configuration.\n" +
				"If it is not found, it will set up a temporary server and database and enter interactive mode.",
			ArgsUsage: "arguments are passed to Lua scripts without parsing",
			Action:    start,
		},
		{
			Name:      "run",
			Usage:     "Run command specified in the project.lua.",
			ArgsUsage: "arguments are passed to Lua scripts without parsing",
			Action:    run,
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		color.Red("%v", err)
	}
}

func start(ctx context.Context, cmd *cli.Command) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	ok, err := config.IsProject(cmd.String("proj"))
	if err != nil {
		return err
	}
	if !ok {
		startInteractive(ctx, cmd)
		return nil
	}
	cfg, err := config.Parse(cmd.String("proj"))
	if err != nil {
		return err
	}
	for k, v := range cfg.Env {
		globalEnv[k] = v
	}

	if cfg.Pid == "" {
		cfg.Pid = path.Join(os.TempDir(), DefaultPidFile)
	}
	if daemon.IsChild() || runtime.GOOS == "windows" {
		return worker(cmd, cfg)
	}

	if err = daemon.WritePid(cfg.Pid); err != nil {
		return err
	}
	defer os.Remove(cfg.Pid)

	ln, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		return err
	}

	var proc, newproc *os.Process
	proc, err = daemon.Fork(wd, ln.(*net.TCPListener))
	if err != nil {
		return err
	}
	if proc == nil {
		warn("warn: running in worker mode\n")
		return worker(cmd, cfg)
	}
	handler := func(sig os.Signal) {
		switch sig {
		case syscall.SIGHUP:
			if err := proc.Kill(); err != nil {
				fail("%v\n", err)
			}
			if newproc, err = daemon.Fork(wd, ln.(*net.TCPListener)); err != nil {
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
		if proc == nil {
			return nil
		}
		_, err := proc.Wait()
		if err != nil {
			return err
		}
	}

	return nil
}

func worker(cmd *cli.Command, cfg config.Config) error {
	sigln := daemon.Listen(daemon.ExitHandler, os.Interrupt)

	ln, err := daemon.Forked(cfg.Listen)
	if err != nil {
		return err
	}

	var capp leapp.Config

	app := fiber.
		New(fiber.Config{
			ServerHeader:          servername,
			DisableStartupMessage: true,
		})
	capp.App = app
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
			c.Set("Server", servername)
			return c.Next()
		})

	apigrp := app.Group(cfg.ApiBase)
	if l := cfg.Limiter; l.Enabled {
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

	admingrp := apigrp.Group(cfg.AdminBase)
	if cfg.Editing {
		admingrp.All("/files/*", admin.Files(cfg.Index))
		warn("warn: editing allowed\n")
	}

	var storage fiber.Storage
	if cfg.DataPath != "" {
		storage = sbolt.New(sbolt.Config{
			Database: path.Join(cfg.DataPath, "fiber.db"),
		})
	}
	capp.Store = session.New(session.Config{
		Storage: storage,
	})

	if cfg.Root != "" {
		ok, err := dir.Is(cfg.Root)
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
			Use(func(ctx *fiber.Ctx) error {
				ctx.Locals(localSkip, strings.HasPrefix(ctx.Path(), cfg.ApiBase))
				return ctx.Next()
			}).
			Use(cache.New(cache.Config{
				Next:         next,
				CacheHeader:  "Cache-Status",
				CacheControl: false,
				Expiration:   72 * time.Hour,
			})).
			Static("/", cfg.Root, fiber.Static{
				Next:      next,
				ByteRange: true,
			})
	}

	capp.Start = func(_ string) error {
		go func() {
			if err := app.Listener(ln); err != nil {
				panic(errors.Wrap(err, "http start"))
			}
		}()
		return nil
	}

	if daemon.IsChild() {
		pid, err := daemon.ReadPid(cfg.Pid)
		if err != nil {
			return err
		}
		capp.Reload = func() error {
			fmt.Println("reloading...")

			if err := daemon.Kill(pid, syscall.SIGHUP); err != nil {
				return err
			}

			os.Exit(0)
			return nil
		}
	}

	capp.Stop = func(timeout time.Duration) error {
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

	G := lue.New(globalEnv)
	defer G.Close()
	G.Register("app", leapp.New(capp)).
		Register("db", ledb.New(cfg.DB)).
		Register("cli", lecli.New(cmd.Args().Slice(), colors)).
		Run(cfg.Index)

	if err := G.Err(); err != nil {
		return err
	}

	sigln.Close()
	if cmd.Bool("interactive") {
		interactive(G)
	} else {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGQUIT)
		<-c
		G.Eval(`app:stop(10)`)
	}

	return nil
}

func startInteractive(ctx context.Context, cmd *cli.Command) {
	fmt.Printf("Mirai Server %s %s\n", version, build)
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
	G := lue.New(globalEnv)
	defer G.Close()
	G.Register("app", leapp.New(capp)).
		Register("db", ledb.New(db)).
		Register("cli", lecli.New(cmd.Args().Slice(), colors))
	if err := G.Err(); err != nil {
		fail("%v\n", err)
	}
	interactive(G)
}

func run(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args()
	if args.Len() < 1 {
		return errors.New("command not specified")
	}

	ok, err := config.IsProject(cmd.String("proj"))
	if err != nil {
		return err
	}

	G := lue.New(globalEnv)
	defer G.Close()

	path := ""
	if ok {
		cfg, err := config.Parse(cmd.String("proj"))
		if err != nil {
			return err
		}
		for k, v := range cfg.Env {
			globalEnv[k] = v
		}
		path, ok = cfg.Commands[args.First()]
		if !ok {
			path = args.First()
		}
		G.Register("db", ledb.New(cfg.DB))
	} else {
		warn("warn: project manifest not found\n")
		path = args.First()
	}
	G.Register("cli", lecli.New(args.Tail(), colors)).
		Run(path)
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
