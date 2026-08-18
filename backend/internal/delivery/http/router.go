package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	customMiddleware "github.com/havilz/caelus-cloud/backend/internal/delivery/http/middleware"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/pkg/config"
)

// NewRouter menginisialisasi router Chi dengan middleware global dan rute inti sistem.
// Parameter cfg merupakan pointer *config.Config yang memuat konfigurasi aplikasi dan origin CORS.
// Mengembalikan pointer *chi.Mux yang siap digunakan sebagai handler HTTP server.
func NewRouter(cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(customMiddleware.CORS(cfg.App.CorsOrigins))
	r.Use(customMiddleware.RequestLogger())

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, http.StatusOK, "Caelus Cloud API is healthy", map[string]string{
			"status":  "ok",
			"service": cfg.App.Name,
			"env":     cfg.App.Env,
		})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			response.Success(w, http.StatusOK, "API v1 is operational", map[string]string{
				"status":  "ok",
				"version": "v1",
			})
		})
	})

	return r
}
