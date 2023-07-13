package leapp

import (
	"log"
	"time"

	"mirai/pkg/lue"

	"github.com/gofiber/contrib/websocket"
	"github.com/vadv/gopher-lua-libs/json"
	lua "github.com/yuin/gopher-lua"
)

// wsAppUpgrade adds a WebSocket handler to the Fiber app.
func wsAppUpgrade(E *lue.Engine) int {
	L := E.L
	app := E.Data(1).(*Application)
	path := L.CheckString(2)
	handler := L.CheckFunction(3)
	p := lua.P{
		Fn:      handler,
		NRet:    0,
		Protect: true,
	}
	wsConnHandler := func(c *websocket.Conn) {
		defer c.Close()

		E, _ := E.New()
		defer E.Close()

		if err := E.L.CallByParam(p, NewWsContext(E, c)); err != nil {
			log.Println()
		}
	}
	app.Use(path, websocket.New(wsConnHandler))
	return 0
}

// NewWsContext creates a new Lua table representing the WebSocket context.
func NewWsContext(E *lue.Engine, c *websocket.Conn) lua.LValue {
	L := E.L
	index := L.NewTable()
	E.SetFields(index, map[string]lua.LValue{
		"headers": E.ReadOnly(mtHttpGetter(c.Headers)),
		"params":  E.ReadOnly(mtHttpGetter(c.Params)),
		"cookies": E.ReadOnly(mtHttpGetter(c.Cookies)),
		"query":   E.ReadOnly(mtHttpGetter(c.Query)),
		"state":   wsCtxState(E, c),
	})
	E.SetFuncs(index, map[string]lue.Fun{
		"send":  wsSend,
		"recv":  wsRecv,
		"ping":  wsPing,
		"close": wsClose,
	})
	return E.Anonymous(c, index)
}

// wsCtxState returns a Lua table representing the WebSocket context's state.
func wsCtxState(E *lue.Engine, c *websocket.Conn) lua.LValue {
	getter := func(key string) lua.LValue {
		if v, ok := c.Locals(key).(lua.LValue); ok {
			return v
		}
		return lua.LNil
	}
	return E.ReadOnly(mtGetter(getter))
}

func wsSend(E *lue.Engine) int {
	L := E.L
	c := E.Data(1).(*websocket.Conn)

	var (
		msg    []byte
		binary bool
	)
	for i := 2; i <= L.GetTop()-1; i++ {
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
}

func wsRecv(E *lue.Engine) int {
	L := E.L
	c := E.Data(1).(*websocket.Conn)

	code, msg, err := c.ReadMessage()
	if err != nil {
		L.RaiseError("ws recv: %v", err)
	}
	L.Push(lua.LNumber(code))
	L.Push(lua.LString(msg))
	return 2
}

func wsPing(E *lue.Engine) int {
	L := E.L
	c := E.Data(1).(*websocket.Conn)

	err := c.WriteControl(websocket.PingMessage, []byte(""), time.Now().Add(5*time.Second))
	if err != nil {
		L.RaiseError("ws ping: %v", err)
	}
	return 0
}

func wsClose(E *lue.Engine) int {
	L := E.L
	c := E.Data(1).(*websocket.Conn)

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
}
