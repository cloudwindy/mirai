package leapp

import (
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/vadv/gopher-lua-libs/json"
	lua "github.com/yuin/gopher-lua"
)

func wsAppUpgrade(a *fiber.App) lua.LGFunction {
	return func(L *lua.LState) int {
		path := L.CheckString(1)
		handler := L.CheckFunction(2)
		p := lua.P{
			Fn:      handler,
			NRet:    0,
			Protect: true,
		}
		wsConnHandler := func(c *websocket.Conn) {
			defer c.Close()
			if err := L.CallByParam(p, NewWsContext(L, c)); err != nil {
				log.Println()
			}
		}
		a.Use(path, websocket.New(wsConnHandler))
		return 0
	}
}

func NewWsContext(L *lua.LState, c *websocket.Conn) lua.LValue {
	ctx := L.NewTable()
	dict := map[string]lua.LValue{
		"headers": objReadOnly(L, mtHttpGetter(c.Headers)),
		"params":  objReadOnly(L, mtHttpGetter(c.Params)),
		"cookies": objReadOnly(L, mtHttpGetter(c.Cookies)),
		"query":   objReadOnly(L, mtHttpGetter(c.Query)),
		"state":   wsCtxState(L, c),
	}
	for k, v := range dict {
		ctx.RawSetString(k, v)
	}
	L.SetFuncs(ctx, wsCtxFuncs(L, c))
	return ctx
}

func wsCtxState(L *lua.LState, c *websocket.Conn) lua.LValue {
	s := c.Locals(localState).(state)
	f := func(key string) lua.LValue {
		if v, ok := s[key]; ok {
			return v
		}
		return lua.LNil
	}
	return objReadOnly(L, mtGetter(f))
}

func wsCtxFuncs(L *lua.LState, c *websocket.Conn) map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"send": func(L *lua.LState) int {
			var (
				msg    []byte
				binary bool
			)
			for i := 1; i <= L.GetTop(); i++ {
				switch param := L.Get(i).(type) {
				case lua.LBool:
					binary = bool(param)
				case lua.LString:
					msg = []byte(param)
				case *lua.LTable:
					msgJson, err := json.ValueEncode(param)
					if err != nil {
						L.RaiseError("ws send json: %v", err)
					}
					msg = msgJson
				default:
					L.RaiseError("ws send: unexpected type %s", param.Type().String())
				}
			}
			msgType := websocket.TextMessage
			if binary {
				msgType = websocket.BinaryMessage
			}
			if err := c.WriteMessage(msgType, msg); err != nil {
				L.RaiseError("ws send: %v", err)
			}
			return 0
		},
		"recv": func(L *lua.LState) int {
			code, msg, err := c.ReadMessage()
			if err != nil {
				L.RaiseError("ws recv: %v", err)
			}
			L.Push(lua.LNumber(code))
			L.Push(lua.LString(msg))
			return 2
		},
		"ping": func(L *lua.LState) int {
			err := c.WriteControl(websocket.PingMessage, []byte(""), time.Now().Add(5*time.Second))
			if err != nil {
				L.RaiseError("ws ping: %v", err)
			}
			return 0
		},
		"close": func(L *lua.LState) int {
			code := websocket.CloseNormalClosure
			text := ""
			if L.GetTop() > 1 {
				code = L.CheckInt(code)
				text = L.CheckString(code)
			} else if L.GetTop() > 0 {
				code = L.CheckInt(code)
			}
			err := c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, text), time.Now().Add(5*time.Second))
			if err != nil {
				L.RaiseError("ws close: %v", err)
			}
			return 0
		},
	}
}
