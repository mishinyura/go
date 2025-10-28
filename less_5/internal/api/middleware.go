package api

import (
	"log"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			d := time.Since(start)
			log.Printf("%s %s -> %v", r.Method, r.URL.Path, d)
		}()
		next.ServeHTTP(w, r)
	})
}
