package url

import (
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/purell"
	lua "github.com/yuin/gopher-lua"
)

var LTURL = "URL"

func New(L *lua.LState) int {
	var u *url.URL
	u = getURL(L, 1)
	if L.GetTop() > 1 {
		base := getURL(L, 2)
		u = base.ResolveReference(u)
	}
	obj := L.NewUserData()
	obj.Value = u
	mt := L.NewTypeMetatable(LTURL)
	if mt.Len() == 0 {
		L.SetFuncs(mt, map[string]lua.LGFunction{
			"__index":    URLGet,
			"__newindex": URLSet,
			"__concat":   URLJoin,
		})
	}
	L.SetMetatable(obj, mt)
	L.Push(obj)
	return 1
}

func Normalize(L *lua.LState) int {
	s := L.CheckString(1)

	s, err := purell.NormalizeURLString(s, purell.FlagsSafe|purell.FlagRemoveDotSegments)
	if err != nil {
		L.RaiseError("url normalize: %v", err)
	}

	L.Push(lua.LString(s))
	return 1
}

func Encode(L *lua.LState) int {
	s := L.CheckString(1)
	s = url.QueryEscape(s)
	L.Push(lua.LString(s))
	return 1
}

func Decode(L *lua.LState) int {
	s := L.CheckString(1)

	s, err := url.QueryUnescape(s)
	if err != nil {
		L.RaiseError("url query unescape: %v", err)
	}

	L.Push(lua.LString(s))
	return 1
}

var URLExports = map[string]lua.LGFunction{
	"resolve": URLJoin,
}

func URLGet(L *lua.LState) int {
	u := L.CheckUserData(1).Value.(*url.URL)
	k := L.CheckString(2)
	switch k {
	case "hash":
		L.Push(lua.LString(u.Fragment))
	case "host":
		L.Push(lua.LString(u.Host))
	case "hostname":
		L.Push(lua.LString(u.Hostname()))
	case "href":
		L.Push(lua.LString(u.String()))
	case "origin":
		L.Push(lua.LString(u.String()))
	case "password":
		if p, ok := u.User.Password(); ok {
			L.Push(lua.LString(p))
		} else {
			L.Push(lua.LNil)
		}
	case "pathname":
		L.Push(lua.LString(u.Path))
	case "port":
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			L.RaiseError("url port: %v", err)
		}
		L.Push(lua.LNumber(port))
	case "protocol":
		L.Push(lua.LString(u.Scheme))
	case "search":
		L.Push(lua.LString(u.Query().Encode()))
	case "search_params":
		q := L.NewTable()
		for k, v := range u.Query() {
			q.RawSetString(k, lua.LString(v[0]))
		}
		L.Push(q)
	case "username":
		L.Push(lua.LString(u.User.Username()))
	default:
		if v, ok := URLExports[k]; ok {
			L.Push(L.NewFunction(v))
		} else {
			L.Push(lua.LNil)
		}
	}
	return 1
}

func URLSet(L *lua.LState) int {
	ud := L.CheckUserData(1)
	k := L.CheckString(2)
	v := L.ToString(3)
	u := ud.Value.(*url.URL)
	switch k {
	case "hash":
		u.Fragment = strings.TrimPrefix(v, "#")
	case "host":
		u.Host = v
	case "hostname":
		_, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			L.RaiseError("url split host port: %v", err)
		}
		u.Host = net.JoinHostPort(v, port)
	case "href":
		u, err := url.Parse(v)
		if err != nil {
			L.RaiseError("url parse: %v", err)
		}
		ud.Value = u
	case "password":
		u.User = url.UserPassword(u.User.Username(), v)
	case "pathname":
		u.Path = v
	case "port":
		hostname, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			L.RaiseError("url split host port: %v", err)
		}
		u.Host = net.JoinHostPort(hostname, v)
	case "protocol":
		u.Scheme = v
	case "search":
		u.RawQuery = v
	case "username":
		p, _ := u.User.Password()
		u.User = url.UserPassword(v, p)
	}
	return 0
}

func URLJoin(L *lua.LState) int {
	ud := L.CheckUserData(1)
	path := L.CheckString(2)

	u := ud.Value.(*url.URL)
	u = u.JoinPath(path)
	ud = L.NewUserData()
	ud.Value = u

	mt := L.GetTypeMetatable(LTURL)
	L.SetMetatable(ud, mt)

	L.Push(ud)
	return 1
}

func getURL(L *lua.LState, n int) *url.URL {
	if L.GetTop() == 0 {
		L.ArgError(1, "expected url")
	}

	v := L.Get(n)
	if ud, ok := v.(*lua.LUserData); ok {
		u, ok := ud.Value.(*url.URL)
		if !ok {
			return nil
		}
		return u
	}
	str := L.CheckString(1)
	if str == "" {
		L.ArgError(1, "url empty")
	}
	str, err := purell.NormalizeURLString(str, purell.FlagsSafe|purell.FlagRemoveDotSegments)
	if err != nil {
		L.RaiseError("url normalize: %v", err)
	}
	u, err := url.Parse(str)
	if err != nil {
		L.RaiseError("url parse: %v", err)
	}
	return u
}
