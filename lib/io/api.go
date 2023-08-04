// Package io implements golang package io functionality for lua.
package io

import (
	"io"

	lio "github.com/vadv/gopher-lua-libs/io"
	lua "github.com/yuin/gopher-lua"
)

func Copy(L *lua.LState) int {
	writer := lio.CheckIOWriter(L, 1)
	reader := lio.CheckIOReader(L, 2)
	var (
		written int64
		err     error
	)
	if L.GetTop() > 2 {
		n := L.CheckInt64(3)
		written, err = io.CopyN(writer, reader, n)
	} else {
		written, err = io.Copy(writer, reader)
	}
	if err != nil {
		L.RaiseError("%v", err)
	}
	L.Push(lua.LNumber(written))
	return 0
}
