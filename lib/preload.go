package lib

import (
	"github.com/cloudwindy/mirai/lib/art"
	"github.com/cloudwindy/mirai/lib/bcrypt"
	"github.com/cloudwindy/mirai/lib/http"
	"github.com/cloudwindy/mirai/lib/io"
	"github.com/cloudwindy/mirai/lib/mail"
	"github.com/cloudwindy/mirai/lib/odbc"
	"github.com/cloudwindy/mirai/lib/os"
	"github.com/cloudwindy/mirai/lib/pwdchecker"
	"github.com/cloudwindy/mirai/lib/re"
	"github.com/cloudwindy/mirai/lib/readline"
	"github.com/cloudwindy/mirai/lib/strings"
	"github.com/cloudwindy/mirai/lib/url"
	"github.com/cloudwindy/mirai/lib/urlpath"
	"github.com/cloudwindy/mirai/lib/uuid"
	"github.com/vadv/gopher-lua-libs/base64"
	"github.com/vadv/gopher-lua-libs/crypto"
	"github.com/vadv/gopher-lua-libs/humanize"
	"github.com/vadv/gopher-lua-libs/inspect"
	"github.com/vadv/gopher-lua-libs/json"
	"github.com/vadv/gopher-lua-libs/storage"
	"github.com/vadv/gopher-lua-libs/time"
	lua "github.com/yuin/gopher-lua"
)

type (
	Load   func(L *lua.LState)
	Loader func(L *lua.LState) int
)

var Loads = []Load{
	io.Load,
	os.Load,
	strings.Load,

	art.Preload,
	bcrypt.Preload,
	http.Preload,
	mail.Preload,
	odbc.Preload,
	pwdchecker.Preload,
	re.Preload,
	readline.Preload,
	uuid.Preload,
	url.Preload,
	urlpath.Preload,

	base64.Preload,
	crypto.Preload,
	humanize.Preload,
	inspect.Preload,
	storage.Preload,
	time.Preload,
}

var Loaders = map[string]Loader{
	"json": json.Loader,
}

func Open(L *lua.LState) {
	for _, load := range Loads {
		load(L)
	}
	for name, loader := range Loaders {
		L.Pop(L.GetTop())
		if loader(L) != 1 {
			panic("loader must return only one value")
		}
		obj := L.Get(1)
		L.SetGlobal(name, obj)
	}
}
