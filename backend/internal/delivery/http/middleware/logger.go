package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

// RequestLogger membuat middleware pencatatan aktivitas HTTP request ke structured logger.
// Mengembalikan closure fungsi middleware func(next http.Handler) http.Handler yang membungkus siklus HTTP request dan mencatat durasi, status HTTP, serta metadata klien.
func RequestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				latency := time.Since(start)
				logger.Info("http_request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"duration_ms", latency.Milliseconds(),
					"remote_ip", r.RemoteAddr,
					"user_agent", r.UserAgent(),
					"bytes_written", ww.BytesWritten(),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
