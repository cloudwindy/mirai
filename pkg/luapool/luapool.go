package luapool

import (
	"sync"

	lua "github.com/yuin/gopher-lua"
)

func New() *LSPool {
	return new(LSPool)
}

type LSPool struct {
	m     sync.Mutex
	Saved []*lua.LState
}

func (pl *LSPool) Get() (L *lua.LState, new bool) {
	pl.m.Lock()
	defer pl.m.Unlock()
	n := len(pl.Saved)
	if n == 0 {
		return pl.New(), true
	}
	x := pl.Saved[n-1]
	pl.Saved = pl.Saved[0 : n-1]
	return x, false
}

func (pl *LSPool) New() *lua.LState {
	L := lua.NewState()
	return L
}

func (pl *LSPool) Put(L *lua.LState) {
	pl.m.Lock()
	defer pl.m.Unlock()
	pl.Saved = append(pl.Saved, L)
}

func (pl *LSPool) Shutdown() {
	for _, L := range pl.Saved {
		L.Close()
	}
}
