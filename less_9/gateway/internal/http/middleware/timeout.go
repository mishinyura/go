package middleware

import (
	"context"
	"net/http"
	"time"
)

const defaultTimeout = 2 * time.Second

func WithTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), defaultTimeout)
		defer cancel()

		r = r.WithContext(ctx)
		done := make(chan struct{})

		go func() {
			next.ServeHTTP(w, r)
			close(done)
		}()

		select {
		case <-ctx.Done():
			http.Error(w, `{"error":"request timeout"}`, http.StatusGatewayTimeout)
		case <-done:
			// все ок
		}
	})
}
