package logging

import (
	"fmt"
	"net/http"
	"time"
)

func (l *Logger) RequestLogger() func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()

			next.ServeHTTP(w, r)

			duration := time.Since(start)

			l.Infoln(
				fmt.Sprintf(
					"%s %s %s %v",
					r.Method,
					r.URL.Path,
					r.RemoteAddr,
					duration,
				),
			)
		})
	}
}
