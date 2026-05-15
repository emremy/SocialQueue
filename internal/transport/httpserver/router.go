package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	// "github.com/emremy/socialqueue/internal/queue/jobs"
)

func CreateRouter(db *gorm.DB, redisClient *redis.Client) http.Handler {
	r := chi.NewRouter()

	// publisher := jobs.NewPublisher(redisClient)

	registerBaseRoutes(r)

	return r
}

func registerBaseRoutes(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}
