package gin

import (
	"github.com/GrygorenkoMykhailo/weir"

	"github.com/gin-gonic/gin"
)

type GinMiddlewareOptions struct {
	Limiter           *weir.Limiter
	OnTooManyRequests gin.HandlerFunc
	KeyExtractor      func(*gin.Context) string
}

func Middleware(weight int, next gin.HandlerFunc, options *GinMiddlewareOptions) gin.HandlerFunc {
	cost := int64(weight)

	return func(ctx *gin.Context) {
		key := options.KeyExtractor(ctx)

		if options.Limiter.Allow(key, cost) {
			next(ctx)
		} else {
			options.OnTooManyRequests(ctx)
		}
	}
}
