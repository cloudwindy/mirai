package lazysess

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type LazySession struct {
	store *session.Store
	s     *session.Session
	c     *fiber.Ctx
}

func (ls *LazySession) Delete(key string) {
	ls.init().Delete(key)
}

func (ls *LazySession) Destroy() error {
	return ls.init().Destroy()
}

func (ls *LazySession) Fresh() bool {
	return ls.init().Fresh()
}

func (ls *LazySession) Get(key string) interface{} {
	return ls.init().Get(key)
}

func (ls *LazySession) ID() string {
	return ls.init().ID()
}

func (ls *LazySession) Keys() []string {
	return ls.init().Keys()
}

func (ls *LazySession) Regenerate() error {
	return ls.init().Regenerate()
}

func (ls *LazySession) Save() error {
	return ls.init().Save()
}

func (ls *LazySession) Set(key string, val interface{}) {
	ls.init().Set(key, val)
}

func (ls *LazySession) SetExpiry(exp time.Duration) {
	ls.init().SetExpiry(exp)
}

func (ls *LazySession) init() Session {
	if ls.s == nil {
		s, err := ls.store.Get(ls.c)
		if err != nil {
			panic(err)
		}
		ls.s = s
	}
	return ls.s
}
