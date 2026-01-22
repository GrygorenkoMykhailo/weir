package stdlib

import (
	"net/http"

	"github.com/GrygorenkoMykhailo/weir"
)

type StdLibMiddlewareOptions struct {
	Limiter           *weir.Limiter
	OnTooManyRequests http.HandlerFunc
	KeyExtractor      func(*http.Request) string
}

func Middleware(weight int, next http.HandlerFunc, options *StdLibMiddlewareOptions) http.HandlerFunc {
	cost := int64(weight)

	return func(w http.ResponseWriter, r *http.Request) {
		key := options.KeyExtractor(r)

		if options.Limiter.Allow(key, cost) {
			next(w, r)
		} else {
			options.OnTooManyRequests(w, r)
		}
	}
}
