package http

import (
	"bytes"
	"io"
	"net/http"

	lua "github.com/yuin/gopher-lua"
)

type luaRequest struct {
	*http.Request
}

const luaRequestType = "http_request_ud"

func checkRequest(L *lua.LState, n int) *luaRequest {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(*luaRequest); ok {
		return v
	}
	L.ArgError(n, "http request expected")
	return nil
}

func lvRequest(L *lua.LState, request *luaRequest) lua.LValue {
	ud := L.NewUserData()
	ud.Value = request
	L.SetMetatable(ud, L.GetTypeMetatable(luaRequestType))
	return ud
}

// http.newreq(verb, url, body) returns request
func NewRequest(L *lua.LState) int {
	verb := L.CheckString(1)
	url := L.CheckString(2)
	var reader io.Reader
	if L.GetTop() > 2 {
		lbody := L.Get(3)
		switch body := lbody.(type) {
		case lua.LString:
			buf := &bytes.Buffer{}
			buf.WriteString(string(body))
			reader = buf
		case *lua.LUserData:
			reader = newLuaFileReader(L, body)
		}
	}
	httpReq, err := http.NewRequest(verb, url, reader)
	if err != nil {
		L.RaiseError("%v", err)
	}

	req := &luaRequest{Request: httpReq}
	req.Request.Header.Set(`User-Agent`, DefaultUserAgent)
	L.Push(lvRequest(L, req))
	return 1
}

// http.req(verb, url, body) returns response
func Request(L *lua.LState) int {
	NewRequest(L)
	req := L.Get(-1)
	cli := GetDefaultClient(L)
	L.Pop(L.GetTop())
	L.Push(cli)
	L.Push(req)
	return DoRequest(L)
}

// request:auth(username, password)
func Auth(L *lua.LState) int {
	req := checkRequest(L, 1)
	user := L.CheckString(2)
	passwd := L.CheckString(3)
	req.SetBasicAuth(user, passwd)
	return 0
}

// request:set(key, value)
func Set(L *lua.LState) int {
	req := checkRequest(L, 1)
	key := L.CheckString(2)
	value := L.CheckString(3)
	req.Header.Set(key, value)
	return 0
}
