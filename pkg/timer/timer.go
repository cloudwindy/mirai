package timer

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Start() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		c.Locals("timer", start)
		return c.Next()
	}
}

func Print(name string, desc string, started ...bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var (
			start    time.Time
			bstarted = !(len(started) != 0 && started[0])
			err      error
		)
		if bstarted {
			start = time.Now()
			c.Locals("timer", start)
			err = c.Next()
		}
		stop := time.Now()
		if start.IsZero() {
			start = c.Locals("timer").(time.Time)
		}
		timing := new(strings.Builder)
		timing.WriteString(name)
		if len(desc) != 0 {
			timing.WriteString(";desc=")
			timing.WriteByte('"')
			timing.WriteString(desc)
			timing.WriteByte('"')
		}
		timing.WriteString(";dur=")
		timing.WriteString(fmt.Sprintf("%.02f", float64(stop.Sub(start).Microseconds())/1000))
		c.Append(fiber.HeaderServerTiming, timing.String())
		if !bstarted {
			err = c.Next()
		}
		return err
	}
}
