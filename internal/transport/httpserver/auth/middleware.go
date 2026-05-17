package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	appauth "github.com/emremy/socialqueue/internal/app/auth"
	response "github.com/emremy/socialqueue/internal/transport/httpserver/response"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

func AuthMiddleware(tokenManager *appauth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractBearerToken(r)

			if tokenString == "" {
				response.WriteJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "missing authorization token",
				})
				return
			}

			claims, err := tokenManager.VerifyAccessToken(tokenString)
			if err != nil {
				response.WriteJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "invalid authorization token",
				})
				return
			}

			userID, err := uuid.Parse(claims.UserID.String())
			if err != nil {
				response.WriteJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "invalid authorization token",
				})
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return userID, ok
}

func extractBearerToken(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))

	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)

	if len(parts) != 2 {
		return ""
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
