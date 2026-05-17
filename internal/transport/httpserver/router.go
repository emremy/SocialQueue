package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	appauth "github.com/emremy/socialqueue/internal/app/auth"

	authhttp "github.com/emremy/socialqueue/internal/transport/httpserver/auth"
	// "github.com/emremy/socialqueue/internal/queue/jobs"
)

func CreateRouter(db *gorm.DB, redisClient *redis.Client, authStore *appauth.Store, tokenManager *appauth.TokenManager) http.Handler {
	r := chi.NewRouter()

	// publisher := jobs.NewPublisher(redisClient)

	registerBaseRoutes(r)

	authService := appauth.NewService(authStore, tokenManager)
	authHandler := authhttp.NewHandler(authService)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)

		r.Group(func(r chi.Router) {
			r.Use(authhttp.AuthMiddleware(tokenManager))

			r.Get("/me", authHandler.Me)
		})
	})

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
