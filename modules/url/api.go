package url

import (
	"net"
	"net/url"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func New(L *lua.LState) int {
	var u *url.URL
	if L.GetTop() == 1 {
		u = GetURL(L, 1)
	} else {
		u = GetURL(L, 2)
		u.Path = L.CheckString(1)
	}
	obj := L.NewUserData()
	obj.Value = u
	mt := L.NewTable()
	L.SetMetatable(obj, mt)
	L.SetFuncs(mt, map[string]lua.LGFunction{
		"__index":    URLGet,
		"__newindex": URLSet,
		"__concat":   URLResolve,
	})
	L.Push(obj)
	return 1
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
		L.Push(lua.LString(u.Port()))
	case "protocol":
		L.Push(lua.LString(u.Scheme))
	case "search":
		L.Push(lua.LString(u.Query().Encode()))
	case "search_params":
		q := L.NewTable()
		for k, v := range u.Query() {
			L.SetField(q, k, lua.LString(v[0]))
		}
		L.Push(q)
	case "username":
		L.Push(lua.LString(u.User.Username()))
	default:
		L.Push(lua.LNil)
	}
	return 1
}

func URLSet(L *lua.LState) int {
	ud := L.CheckUserData(1)
	k := L.CheckString(2)
	v := L.CheckString(3)
	u := ud.Value.(*url.URL)
	switch k {
	case "hash":
		u.Fragment = strings.TrimPrefix(v, "#")
	case "host":
		u.Host = v
	case "hostname":
		_, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			L.Error(lua.LString(err.Error()), 1)
		}
		u.Host = net.JoinHostPort(v, port)
	case "href":
		u, err := url.Parse(v)
		if err != nil {
			L.Error(lua.LString(err.Error()), 1)
		}
		ud.Value = u
	case "password":
		u.User = url.UserPassword(u.User.Username(), v)
	case "pathname":
		u.Path = v
	case "port":
		hostname, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			L.Error(lua.LString(err.Error()), 1)
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

func URLResolve(L *lua.LState) int {
	ud := L.CheckUserData(1)
	path := L.CheckString(2)
	u := ud.Value.(*url.URL)
	path = u.JoinPath(path).Path
	L.Push(lua.LString(path))
	return 1
}

func GetURL(L *lua.LState, n int) *url.URL {
	if L.GetTop() == 0 {
		L.Error(lua.LString("expected url"), 1)
	}
	ud := L.OptUserData(n, nil)
	if ud == nil {
		u, ok := ud.Value.(*url.URL)
		if !ok {
			return nil
		}
		return u
	}
	str := L.OptString(n, "")
	if str == "" {
		L.Error(lua.LString("url empty"), 1)
	}
	u, err := url.Parse(str)
	if err != nil {
		L.Error(lua.LString(err.Error()), 1)
	}
	return u
}
