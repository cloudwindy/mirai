package libs

import (
	"mirai/lib/modules/bcrypt"
	"mirai/lib/modules/mail"
	"mirai/lib/modules/pwdchecker"
	"mirai/lib/modules/url"
	"mirai/lib/modules/urlpath"

	"github.com/vadv/gopher-lua-libs/base64"
	"github.com/vadv/gopher-lua-libs/chef"
	"github.com/vadv/gopher-lua-libs/cmd"
	"github.com/vadv/gopher-lua-libs/crypto"
	"github.com/vadv/gopher-lua-libs/filepath"
	http_client "github.com/vadv/gopher-lua-libs/http/client"
	"github.com/vadv/gopher-lua-libs/humanize"
	"github.com/vadv/gopher-lua-libs/inspect"
	"github.com/vadv/gopher-lua-libs/json"
	"github.com/vadv/gopher-lua-libs/log"
	"github.com/vadv/gopher-lua-libs/pb"
	"github.com/vadv/gopher-lua-libs/regexp"
	"github.com/vadv/gopher-lua-libs/runtime"
	"github.com/vadv/gopher-lua-libs/stats"
	"github.com/vadv/gopher-lua-libs/storage"
	"github.com/vadv/gopher-lua-libs/time"
	lua "github.com/yuin/gopher-lua"
)

func PreloadAll(L *lua.LState) {
	bcrypt.Preload(L)
	mail.Preload(L)
	pwdchecker.Preload(L)
	url.Preload(L)
	urlpath.Preload(L)

	base64.Preload(L)
	chef.Preload(L)
	cmd.Preload(L)
	crypto.Preload(L)
	filepath.Preload(L)
	humanize.Preload(L)
	inspect.Preload(L)
	json.Preload(L)
	log.Preload(L)
	pb.Preload(L)
	runtime.Preload(L)
	stats.Preload(L)
	storage.Preload(L)
	time.Preload(L)
	L.PreloadModule("http", http_client.Loader)
	L.PreloadModule("re", regexp.Loader)
}
