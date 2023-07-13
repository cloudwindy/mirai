package lutpool

import (
	"sync"

	lua "github.com/yuin/gopher-lua"
)

func New(opt ...lua.Options) *LSPool {
	pool := new(LSPool)
	if len(opt) > 0 {
		pool.options = opt[0]
	}
	return pool
}

type LSPool struct {
	m       sync.Mutex
	options lua.Options
	Saved   []*lua.LState
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
	return lua.NewState(pl.options)
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
