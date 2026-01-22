package fiber

import (
	"github.com/GrygorenkoMykhailo/weir"

	"github.com/gofiber/fiber/v2"
)

type FiberMiddlewareOptions struct {
	Limiter           *weir.Limiter
	OnTooManyRequests fiber.Handler
	KeyExtractor      func(*fiber.Ctx) string
}

func Middleware(weight int, next fiber.Handler, options *FiberMiddlewareOptions) fiber.Handler {
	cost := int64(weight)

	return func(c *fiber.Ctx) error {
		key := options.KeyExtractor(c)

		if options.Limiter.Allow(key, cost) {
			return next(c)
		}

		return options.OnTooManyRequests(c)
	}
}
