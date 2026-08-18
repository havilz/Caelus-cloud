package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORS mengonfigurasi dan mengembalikan middleware Cross-Origin Resource Sharing (CORS).
// Parameter allowedOrigins merupakan slice string domain origin yang diberikan izin akses ke API.
// Mengembalikan fungsi middleware func(next http.Handler) http.Handler yang menyisipkan header CORS pada setiap response.
func CORS(allowedOrigins []string) func(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
