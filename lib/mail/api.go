package mail

import (
	"net/url"
	"strings"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	lua "github.com/yuin/gopher-lua"
)

func Send(L *lua.LState) int {
	uri := L.CheckString(1)
	from := L.CheckString(2)
	luaTo := L.CheckTable(3)
	body := L.CheckString(4)

	u, err := url.Parse(uri)
	if err != nil {
		L.RaiseError("invalid uri: %v", err)
	}
	addr := u.Host
	username := u.User.Username()
	password, _ := u.User.Password()
	to := make([]string, 0)
	luaTo.ForEach(func(l1, l2 lua.LValue) {
		to = append(to, l2.String())
	})

	c, err := smtp.Dial(addr)
	if err != nil {
		L.RaiseError("smtp dial failed: %v", err)
	}

	if u.Scheme == "smtps" {
		if err := c.StartTLS(nil); err != nil {
			L.RaiseError("smtp starttls failed: %v", err)
		}
	}

	if username != "" && password != "" {
		sc := sasl.NewPlainClient("", username, password)
		if err := c.Auth(sc); err != nil {
			L.RaiseError("smtp auth failed: %v", err)
		}
	}

	err = c.SendMail(from, to, strings.NewReader(body))
	if err != nil {
		L.RaiseError("smtp send failed: %v", err)
	}
	return 0
}
