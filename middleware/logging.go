package middleware

import (
	"log"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("\033[44m %s \033[0m | PATH: \033[33m\"%s\"\033[0m | TIME: %v",
			r.Method, r.URL.Path, start.Format("2006-01-02 15:04:05"),
		)
		next.ServeHTTP(w, r)
	})
}
