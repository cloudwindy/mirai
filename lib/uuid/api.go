package uuid

import (
	"github.com/google/uuid"
	lua "github.com/yuin/gopher-lua"
)

func New(L *lua.LState) int {
	L.Push(lua.LString(uuid.New().String()))
	return 1
}

func ToBytes(L *lua.LState) int {
	u := L.CheckString(1)
	bin, err := uuid.MustParse(u).MarshalBinary()
	if err != nil {
		L.RaiseError("tobytes: %v", err)
	}
	L.Push(lua.LString(bin))
	return 1
}

func FromBytes(L *lua.LState) int {
	u := L.CheckString(1)
	id, err := uuid.FromBytes([]byte(u))
	if err != nil {
		L.RaiseError("frombytes: %v", err)
	}
	L.Push(lua.LString(id.String()))
	return 1
}