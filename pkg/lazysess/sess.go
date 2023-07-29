package lazysess

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type Session interface {
	Delete(key string)
	Destroy() error
	Fresh() bool
	Get(key string) any
	ID() string
	Keys() []string
	Regenerate() error
	Save() error
	Set(key string, val any)
	SetExpiry(exp time.Duration)
}

func New(c *fiber.Ctx, store *session.Store) Session {
	return &LazySession{c: c, store: store}
}
