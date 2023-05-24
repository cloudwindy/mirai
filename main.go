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

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	sbolt "github.com/gofiber/storage/bbolt"
	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
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

	fmt.Println(art.String("Mirai Project"))
	fmt.Println(" Mirai Server " + Version + " with " + lua.LuaVersion)
	fmt.Println(" Fiber " + fiber.Version)

	app := fiber.New(fiber.Config{
		ServerHeader:          "Mirai Server",
		DisableStartupMessage: true,
	})
	app.Use(logger.New())
	app.Use(recover.New())
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
		Next:        skipApi,
		CacheHeader: "Cache-Status",
		Expiration:  24 * time.Hour,
	}))

	db, err := bolt.Open("db/data.db", 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		panic(err)
	}

	storage := sbolt.New(sbolt.Config{
		Database: "db/sessions.db",
	})
	store := session.New(session.Config{
		Storage: storage,
	})

	api := app.Group("/api")
	api.All("/v1/:scriptname?/*", luaHandler("./v1", db, store))
	admin := api.Group("/admin")
	if EnableEditing {
		fmt.Println(" Editing: Enabled")
		admin.All("/files/:path?", filesHandler("./v1"))
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
			fmt.Println("cache missed")
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
			return err
		}

		return nil
	}
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

	L.SetField(req, "base_url", lua.LString(c.BaseURL()))
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
	L.SetField(sess, "keys", L.NewFunction(func(L *lua.LState) int {
		k := L.NewTable()
		for _, key := range s.Keys() {
			k.Append(lua.LString(key))
		}
		L.Push(k)
		return 1
	}))
	L.SetField(sess, "save", L.NewFunction(func(l *lua.LState) int {
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

	mtSess := L.NewTable()
	L.SetField(mtSess, "__index", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(2)
		value, ok := s.Get(key).(string)
		if !ok {
			L.Push(lua.LNil)
		}
		L.Push(lua.LString(value))
		return 1
	}))
	L.SetField(mtSess, "__newindex", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(2)
		value := L.ToString(3)
		s.Set(key, value)
		return 0
	}))
	L.SetMetatable(sess, mtSess)
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
  L.SetField(mt, "__index", L.NewFunction(lkvIndex))
  L.SetField(mt, "__newindex", L.NewFunction(lkvPut))
}

var lkvExports = map[string]lua.LGFunction{
  "__index": lkvGet,
  "__newindex": lkvPut,
  "keys": lkvKeys,
  "drop": lkvDrop,
}

func lkvIndex(L *lua.LState) int {
  switch L.CheckString(2) {
    case "keys":
      L.Push(L.NewFunction(lkvKeys))
      return 1
    case "drop":
      L.Push(L.NewFunction(lkvDrop))
      return 1
    default:
      return lkvGet(L)
  }
}

func lkvCheck(L *lua.LState) *Bucket {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Bucket); ok {
		return v
	}
	L.ArgError(1, "bucket expected")
	return nil
}

func lkvGet(L *lua.LState) int {
	bucket := lkvCheck(L)
	key := L.CheckString(2)
	res, err := kvGet(bucket.DB, bucket.Name, key)
	if err != nil {
		L.Error(lua.LString(err.Error()), 1)
	}
	if res == nil {
		L.Push(lua.LNil)
	}
	L.Push(lua.LString(*res))
	return 1
}

func lkvKeys(L *lua.LState) int {
	bucket := lkvCheck(L)
	res, err := kvKeys(bucket.DB, bucket.Name)
	if err != nil {
		L.Error(lua.LString(err.Error()), 1)
	}
	if res == nil {
		L.Push(lua.LNil)
		return 1
	}
	t := L.NewTable()
	for _, v := range res {
		t.Append(lua.LString(v))
	}
	L.Push(t)
	return 1
}

func lkvPut(L *lua.LState) int {
	bucket := lkvCheck(L)
	key := L.CheckString(2)
	value := L.CheckString(3)
	if err := kvPut(bucket.DB, bucket.Name, key, value); err != nil {
		L.Error(lua.LString(err.Error()), 1)
	}
	return 0
}

func lkvDrop(L *lua.LState) int {
	bucket := lkvCheck(L)
	if err := kvDrop(bucket.DB, bucket.Name); err != nil {
		L.Error(lua.LString(err.Error()), 1)
	}
	return 0
}

func kvBucket(tx *bolt.Tx, bucket []string) *bolt.Bucket {
	b := tx.Bucket([]byte(bucket[0]))
	if b == nil {
		return nil
	}
	for _, part := range bucket[1:] {
		b = b.Bucket([]byte(part))
		if b == nil {
			return nil
		}
	}
	return b
}

var delimiter = ":"

func kvGet(db *bolt.DB, bucket, key string) (*string, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	parts := strings.Split(bucket, delimiter)
	b := kvBucket(tx, parts)
	res := string(b.Get([]byte(key)))
	return &res, nil
}

func kvKeys(db *bolt.DB, bucket string) ([]string, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	parts := strings.Split(bucket, delimiter)
	b := kvBucket(tx, parts)
	res := make([]string, 0)
	err = b.ForEach(func(k, v []byte) error {
		res = append(res, string(k))
		return nil
	})
	return res, err
}

func kvPut(db *bolt.DB, bucket, key, value string) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	parts := strings.Split(bucket, delimiter)
	b := kvBucket(tx, parts)
	err = b.Put([]byte(key), []byte(value))
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func kvDrop(db *bolt.DB, bucket string) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	parts := strings.Split(bucket, delimiter)
	if len(parts) <= 1 {
		if err = tx.DeleteBucket([]byte(bucket)); err != nil {
			return err
		}
	} else {
		b := kvBucket(tx, parts[:len(parts)-2])
		if err = b.DeleteBucket([]byte(bucket)); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
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

func hardStop(termCh chan os.Signal, stopCh chan any) {
	select {
	case <-termCh:
		// terminate
		os.Exit(1)
	case <-stopCh:
		return
	}
}
