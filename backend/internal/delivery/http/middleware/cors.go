package middleware

import (
	"net/http"
	"strings"

	"github.com/go-chi/cors"
)

func CORS(allowedOrigins []string) func(next http.Handler) http.Handler {
	opts := cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID", "X-Server-ID"},
		ExposedHeaders:   []string{"Link", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	allowAll := false
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAll = true
			break
		}
	}

	if allowAll || len(allowedOrigins) == 0 {
		opts.AllowOriginFunc = func(_ *http.Request, _ string) bool {
			return true
		}
	} else {
		opts.AllowOriginFunc = func(_ *http.Request, origin string) bool {
			for _, o := range allowedOrigins {
				if o == origin {
					return true
				}
			}

			if strings.Contains(origin, "localhost") ||
				strings.Contains(origin, "127.0.0.1") ||
				strings.Contains(origin, "trycloudflare.com") ||
				strings.Contains(origin, "100.") ||
				strings.Contains(origin, "192.168.") ||
				strings.Contains(origin, "10.") ||
				strings.Contains(origin, "172.") {
				return true
			}
			return false
		}
	}

	return cors.Handler(opts)
}
