package main

import (
	"github.com/vadv/gopher-lua-libs/goos"
	"github.com/vadv/gopher-lua-libs/ioutil"
	"github.com/vadv/gopher-lua-libs/strings"
	lua "github.com/yuin/gopher-lua"
)

func OpenExtendLib(L *lua.LState) {
	modString := L.RegisterModule("string", nil).(*lua.LTable)
	L.SetFuncs(modString, stringExports)
	modIo := L.RegisterModule("io", nil).(*lua.LTable)
	L.SetFuncs(modIo, ioExports)
	modOs := L.RegisterModule("os", nil).(*lua.LTable)
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
