// Package io implements golang package io functionality for lua.
package io

import (
	"io"
	"os"

	lio "github.com/vadv/gopher-lua-libs/io"

	lua "github.com/yuin/gopher-lua"
)

// ReadFile lua io.read_file(filepath) reads the file named by filename and returns the contents, returns (string,error)
func ReadFile(L *lua.LState) int {
	filename := L.CheckString(1)
	data, err := os.ReadFile(filename)
	if err != nil {
		L.RaiseError("%v", err)
	}
	L.Push(lua.LString(data))
	return 1
}

// WriteFile lua io.write_file(filepath, data) reads the file named by filename and returns the contents, returns (string,error)
func WriteFile(L *lua.LState) int {
	filename := L.CheckString(1)
	data := L.CheckString(2)
	err := os.WriteFile(filename, []byte(data), 0o644)
	if err != nil {
		L.RaiseError("%v", err)
	}
	return 0
}

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
	}
	if err != nil {
		L.RaiseError("%v", err)
	}
	L.Push(lua.LNumber(written))
	return 0
}
