package lelib

import (
	"mirai/lib/bcrypt"
	"mirai/lib/mail"
	"mirai/lib/pwdchecker"
	"mirai/lib/readline"
	"mirai/lib/url"
	"mirai/lib/urlpath"
	"mirai/lib/uuid"

	"github.com/vadv/gopher-lua-libs/base64"
	"github.com/vadv/gopher-lua-libs/cmd"
	"github.com/vadv/gopher-lua-libs/crypto"
	"github.com/vadv/gopher-lua-libs/filepath"
	"github.com/vadv/gopher-lua-libs/goos"
	http "github.com/vadv/gopher-lua-libs/http/client"
	"github.com/vadv/gopher-lua-libs/humanize"
	"github.com/vadv/gopher-lua-libs/inspect"
	"github.com/vadv/gopher-lua-libs/ioutil"
	"github.com/vadv/gopher-lua-libs/json"
	"github.com/vadv/gopher-lua-libs/log"
	"github.com/vadv/gopher-lua-libs/regexp"
	"github.com/vadv/gopher-lua-libs/runtime"
	"github.com/vadv/gopher-lua-libs/storage"
	"github.com/vadv/gopher-lua-libs/strings"
	"github.com/vadv/gopher-lua-libs/time"
	lua "github.com/yuin/gopher-lua"
)

func OpenLib(L *lua.LState) {
	PreloadAll(L)
	modString := L.RegisterModule(lua.StringLibName, nil).(*lua.LTable)
	L.SetFuncs(modString, stringExports)
	modIo := L.RegisterModule(lua.IoLibName, nil).(*lua.LTable)
	L.SetFuncs(modIo, ioExports)
	modOs := L.RegisterModule(lua.OsLibName, nil).(*lua.LTable)
	L.SetFuncs(modOs, osExports)
}

func PreloadAll(L *lua.LState) {
	bcrypt.Preload(L)
	mail.Preload(L)
	pwdchecker.Preload(L)
	readline.Preload(L)
	uuid.Preload(L)
	url.Preload(L)
	urlpath.Preload(L)

	base64.Preload(L)
	cmd.Preload(L)
	crypto.Preload(L)
	filepath.Preload(L)
	humanize.Preload(L)
	inspect.Preload(L)
	json.Preload(L)
	log.Preload(L)
	runtime.Preload(L)
	storage.Preload(L)
	time.Preload(L)
	L.PreloadModule("http", http.Loader)
	L.PreloadModule("re", regexp.Loader)
}

var stringExports = map[string]lua.LGFunction{
	"split":      strings.Split,
	"fields":     strings.Fields,
	"includes":   strings.Contains,
	"startswith": strings.HasPrefix,
	"endswith":   strings.HasSuffix,
	"trim":       strings.Trim,
	"trimspace":  strings.TrimSpace,
	"trimstart":  strings.TrimPrefix,
	"trimend":    strings.TrimSuffix,
}

var ioExports = map[string]lua.LGFunction{
	"read_file":  ioutil.ReadFile,
	"write_file": ioutil.WriteFile,
	"copy":       ioutil.Copy,
	"copyn":      ioutil.CopyN,
}

var osExports = map[string]lua.LGFunction{
	"stat":         goos.Stat,
	"hostname":     goos.Hostname,
	"get_pagesize": goos.Getpagesize,
	"mkdir_all":    goos.MkdirAll,
}
