package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.status != 0 {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(resp []byte) (int, error) {
	if rw.status == 0 {
		rw.WriteHeader(http.StatusOK)
	}

	res, err := rw.ResponseWriter.Write(resp)
	rw.bytes += res

	return res, err
}

func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{
				ResponseWriter: w,
			}

			next.ServeHTTP(rw, r)

			if rw.status == 0 {
				rw.status = http.StatusOK
			}

			logger.Info(
				"request_completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"request_id", middleware.GetReqID(r.Context()),
				"bytes", rw.bytes,
				"duration", time.Since(start),
			)
		})
	}
}
