package stdlib

import (
	"net/http"
	"strconv"

	"github.com/tinycolds/limiter"
)

// Middleware is the middleware for basic http.Handler.
type Middleware struct {
	Limiter            *limiter.Limiter
	OnError            ErrorHandler
	OnLimitReached     LimitReachedHandler
	TrustForwardHeader bool
	GlobalLimit        bool
}

// NewMiddleware return a new instance of a basic HTTP middleware.
func NewMiddleware(limiter *limiter.Limiter, options ...Option) *Middleware {
	middleware := &Middleware{
		Limiter:        limiter,
		OnError:        DefaultErrorHandler,
		OnLimitReached: DefaultLimitReachedHandler,
	}

	for _, option := range options {
		option.Apply(middleware)
	}

	return middleware
}

// Handler the middleware handler.
func (middleware *Middleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var context 	limiter.Context
		var err 	error

		if middleware.GlobalLimit {
			context, err = middleware.Limiter.Get(r.Context(), limiter.GetDefaultKey(r))
		} else {
			context, err = middleware.Limiter.Get(r.Context(), limiter.GetIPKey(r, middleware.TrustForwardHeader))
		}

		if err != nil {
			middleware.OnError(w, r, err)
			return
		}

		w.Header().Add("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		w.Header().Add("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		w.Header().Add("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			middleware.OnLimitReached(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}
