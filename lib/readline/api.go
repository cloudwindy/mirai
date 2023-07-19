package readline

import (
	"errors"

	"github.com/chzyer/readline"
	lua "github.com/yuin/gopher-lua"
)

func Readline(L *lua.LState) int {
	prompt := L.CheckString(1)
	line, err := readline.Line(prompt)
	if err != nil {
		if !errors.Is(err, readline.ErrInterrupt) {
			L.RaiseError("readline read: %v", err)
		}
		L.Push(lua.LNil)
	} else {
		L.Push(lua.LString(line))
	}
	return 1
}

func AddHistory(L *lua.LState) int {
	content := L.CheckString(1)
	if err := readline.AddHistory(content); err != nil {
		L.RaiseError("readline addhistory: %v", err)
	}
	return 0
}
