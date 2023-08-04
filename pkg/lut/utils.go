package lut

import (
	"bufio"
	"fmt"
	"os"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

// CompileLua reads the passed lua file from disk and compiles it.
func CompileLua(filePath string) (*lua.FunctionProto, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	chunk, err := parse.Parse(reader, filePath)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, err
	}
	if err := CheckGlobal(proto, filePath); err != nil {
		return nil, err
	}
	return proto, nil
}

func DoCompiledFile(L *lua.LState, proto *lua.FunctionProto) error {
	lfunc := L.NewFunctionFromProto(proto)
	L.Push(lfunc)
	return L.PCall(0, lua.MultRet, nil)
}

func Unprotect(fn *lua.LFunction, self lua.LValue, nret int) lua.LGFunction {
	return func(L *lua.LState) int {
		L.Replace(1, self)
		L.Insert(fn, 1)
		L.Call(L.GetTop()-1, nret)
		err := L.Get(-1)
		if err != lua.LNil {
			L.RaiseError(lua.LVAsString(err))
		}
		for i := 1; i <= nret-1; i++ {
			L.Push(L.Get(i))
		}
		return nret - 1
	}
}

func CheckGlobal(proto *lua.FunctionProto, source string) error {
	for i, code := range proto.Code {
		if opGetOpCode(code) == lua.OP_SETGLOBAL {
			pos := proto.DbgSourcePositions[i]
			return fmt.Errorf("compile error near line(%v) %v: %v", pos, source, "variable not defined")
		}
	}
	for _, nestedProto := range proto.FunctionPrototypes {
		if err := CheckGlobal(nestedProto, source); err != nil {
			return err
		}
	}
	return nil
}

func opGetOpCode(inst uint32) int {
	return int(inst >> 26)
}
