package luapool

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"mirai/pkg/luaextlib"

	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

type StatePool struct {
	m     sync.Mutex
	Saved []*lua.LState
}

func (pl *StatePool) Get() *lua.LState {
	pl.m.Lock()
	defer pl.m.Unlock()
	n := len(pl.Saved)
	if n == 0 {
		return pl.New()
	}
	x := pl.Saved[n-1]
	pl.Saved = pl.Saved[0 : n-1]
	return x
}

func (pl *StatePool) New() *lua.LState {
	L := lua.NewState()
	// setting the L up here.
	// load scripts, set global variables, share channels, etc...
	luaextlib.OpenLib(L)
	return L
}

func (pl *StatePool) Put(L *lua.LState) {
	pl.m.Lock()
	defer pl.m.Unlock()
	pl.Saved = append(pl.Saved, L)
}

func (pl *StatePool) Shutdown() {
	for _, L := range pl.Saved {
		L.Close()
	}
}

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

func CheckGlobal(proto *lua.FunctionProto, source string) error {
	for i, code := range proto.Code {
		if opGetOpCode(code) == lua.OP_SETGLOBAL {
			pos := proto.DbgSourcePositions[i]
			return errors.New(fmt.Sprintf("compile error near line(%v) %v: %v", pos, source, "variable not defined"))
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
