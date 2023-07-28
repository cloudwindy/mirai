package leapp

import (
	"log"
	"time"

	"github.com/cloudwindy/mirai/pkg/lue"
	"github.com/gofiber/contrib/websocket"
	"github.com/vadv/gopher-lua-libs/json"
	lua "github.com/yuin/gopher-lua"
)

// wsAppUpgrade adds a WebSocket handler to the Fiber app.
func wsAppUpgrade(E *lue.Engine) int {
	app := E.Data(1).(*Application)
	path := E.String(2)
	handler := E.Fun(3)
	wsConnHandler := func(c *websocket.Conn) {
		defer c.Close()
		E, _ := E.New()
		defer E.Close()
		if err := E.CallLFun(handler, 0, NewWsContext(E, c)); err != nil {
			log.Println()
		}
	}
	app.Use(path, websocket.New(wsConnHandler))
	return 0
}

// NewWsContext creates a new Lua table representing the WebSocket context.
func NewWsContext(E *lue.Engine, c *websocket.Conn) lua.LValue {
	index := E.NewTable()

	E.SetFields(index, map[string]lua.LValue{
		"headers": E.ReadOnly(mtHttpGetter(c.Headers)),
		"params":  E.ReadOnly(mtHttpGetter(c.Params)),
		"cookies": E.ReadOnly(mtHttpGetter(c.Cookies)),
		"query":   E.ReadOnly(mtHttpGetter(c.Query)),
		"state":   wsCtxState(E, c),
	})
	E.SetFuncs(index, wsExports)

	return E.Anonymous(c, index)
}

var wsExports = map[string]lue.Fun{
	"send":  wsSend,
	"recv":  wsRecv,
	"ping":  wsPing,
	"close": wsClose,
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
	c := E.Data(1).(*websocket.Conn)

	var (
		msg    []byte
		binary bool
	)
	for i := 2; i <= E.Top()-1; i++ {
		switch param := E.Get(i).(type) {
		case lua.LBool:
			binary = bool(param)
		case lua.LString:
			msg = []byte(param)
		case *lua.LTable:
			msgJson, err := json.ValueEncode(param)
			if err != nil {
				E.Error("ws send json: %v", err)
			}
			msg = msgJson
		default:
			E.Error("ws send: unexpected type %s", param.Type().String())
		}
	}
	msgType := websocket.TextMessage
	if binary {
		msgType = websocket.BinaryMessage
	}
	if err := c.WriteMessage(msgType, msg); err != nil {
		E.Error("ws send: %v", err)
	}
	return 0
}

func wsRecv(E *lue.Engine) int {
	c := E.Data(1).(*websocket.Conn)

	code, msg, err := c.ReadMessage()
	if err != nil {
		E.Error("ws recv: %v", err)
	}
	E.PushInt(code)
	E.PushString(string(msg))
	return 2
}

func wsPing(E *lue.Engine) int {
	c := E.Data(1).(*websocket.Conn)

	err := c.WriteControl(websocket.PingMessage, []byte(""), time.Now().Add(5*time.Second))
	if err != nil {
		E.Error("ws ping: %v", err)
	}
	return 0
}

func wsClose(E *lue.Engine) int {
	c := E.Data(1).(*websocket.Conn)

	code := websocket.CloseNormalClosure
	text := ""
	if E.Top() > 1 {
		code = E.Int(code)
		text = E.String(code)
	} else if E.Top() > 0 {
		code = E.Int(code)
	}
	err := c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, text), time.Now().Add(5*time.Second))
	if err != nil {
		E.Error("ws close: %v", err)
	}
	return 0
}
