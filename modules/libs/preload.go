package libs

import (
	"mirai/modules/bcrypt"
	"mirai/modules/pwdchecker"

	"github.com/vadv/gopher-lua-libs/base64"
	"github.com/vadv/gopher-lua-libs/chef"
	"github.com/vadv/gopher-lua-libs/cmd"
	"github.com/vadv/gopher-lua-libs/crypto"
	"github.com/vadv/gopher-lua-libs/filepath"
	"github.com/vadv/gopher-lua-libs/goos"
	"github.com/vadv/gopher-lua-libs/http"
	"github.com/vadv/gopher-lua-libs/humanize"
	"github.com/vadv/gopher-lua-libs/inspect"
	"github.com/vadv/gopher-lua-libs/ioutil"
	"github.com/vadv/gopher-lua-libs/json"
	"github.com/vadv/gopher-lua-libs/log"
	"github.com/vadv/gopher-lua-libs/pb"
	"github.com/vadv/gopher-lua-libs/regexp"
	"github.com/vadv/gopher-lua-libs/runtime"
	"github.com/vadv/gopher-lua-libs/stats"
	"github.com/vadv/gopher-lua-libs/storage"
	"github.com/vadv/gopher-lua-libs/strings"
	"github.com/vadv/gopher-lua-libs/time"
	lua "github.com/yuin/gopher-lua"
)

func PreloadAll(L *lua.LState) {
	bcrypt.Preload(L)
	pwdchecker.Preload(L)

	base64.Preload(L)
	chef.Preload(L)
	cmd.Preload(L)
	crypto.Preload(L)
	filepath.Preload(L)
	goos.Preload(L)
	http.Preload(L)
	humanize.Preload(L)
	inspect.Preload(L)
	ioutil.Preload(L)
	json.Preload(L)
	log.Preload(L)
	pb.Preload(L)
	regexp.Preload(L)
	runtime.Preload(L)
	stats.Preload(L)
	storage.Preload(L)
	strings.Preload(L)
	time.Preload(L)
}
