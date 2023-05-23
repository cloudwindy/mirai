package bcrypt

import (
	"errors"

	lua "github.com/yuin/gopher-lua"
	"golang.org/x/crypto/bcrypt"
)

func Hash(L *lua.LState) int {
	password := L.CheckString(1)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(hash))
	return 1
}

func Compare(L *lua.LState) int {
	hash := L.CheckString(1)
	password := L.CheckString(2)
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			L.Push(lua.LFalse)
			return 1
		}
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 0
	}
	L.Push(lua.LTrue)
	return 1
}
