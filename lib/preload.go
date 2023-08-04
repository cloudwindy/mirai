package lib

import (
	"github.com/cloudwindy/mirai/lib/art"
	"github.com/cloudwindy/mirai/lib/bcrypt"
	"github.com/cloudwindy/mirai/lib/http"
	"github.com/cloudwindy/mirai/lib/io"
	"github.com/cloudwindy/mirai/lib/mail"
	"github.com/cloudwindy/mirai/lib/odbc"
	"github.com/cloudwindy/mirai/lib/pwdchecker"
	"github.com/cloudwindy/mirai/lib/re"
	"github.com/cloudwindy/mirai/lib/readline"
	"github.com/cloudwindy/mirai/lib/url"
	"github.com/cloudwindy/mirai/lib/urlpath"
	"github.com/cloudwindy/mirai/lib/uuid"
	"github.com/vadv/gopher-lua-libs/base64"
	"github.com/vadv/gopher-lua-libs/cmd"
	"github.com/vadv/gopher-lua-libs/crypto"
	"github.com/vadv/gopher-lua-libs/filepath"
	"github.com/vadv/gopher-lua-libs/goos"
	"github.com/vadv/gopher-lua-libs/humanize"
	"github.com/vadv/gopher-lua-libs/inspect"
	"github.com/vadv/gopher-lua-libs/json"
	"github.com/vadv/gopher-lua-libs/log"
	"github.com/vadv/gopher-lua-libs/runtime"
	"github.com/vadv/gopher-lua-libs/storage"
	"github.com/vadv/gopher-lua-libs/strings"
	"github.com/vadv/gopher-lua-libs/time"
	lua "github.com/yuin/gopher-lua"
)

func Preload(L *lua.LState) {
	art.Preload(L)
	bcrypt.Preload(L)
	http.Preload(L)
	mail.Preload(L)
	odbc.Preload(L)
	pwdchecker.Preload(L)
	re.Preload(L)
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
}

func Open(L *lua.LState) {
	Preload(L)
	modString := L.RegisterModule(lua.StringLibName, nil).(*lua.LTable)
	L.SetFuncs(modString, stringExports)
	modIo := L.RegisterModule(lua.IoLibName, nil).(*lua.LTable)
	L.SetFuncs(modIo, ioExports)
	modOs := L.RegisterModule(lua.OsLibName, nil).(*lua.LTable)
	L.SetFuncs(modOs, osExports)
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
	"readfile":  io.ReadFile,
	"writefile": io.WriteFile,
	"copy":      io.Copy,
}

var osExports = map[string]lua.LGFunction{
	"stat":         goos.Stat,
	"hostname":     goos.Hostname,
	"get_pagesize": goos.Getpagesize,
	"mkdir":        goos.MkdirAll,
}
