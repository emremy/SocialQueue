package auth

import (
	"time"

	appauth "github.com/emremy/socialqueue/internal/app/auth"
	"github.com/google/uuid"
)

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FullName  *string   `json:"full_name,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenResponse struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

type AuthResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

func NewAuthResponse(result appauth.AuthResult) AuthResponse {
	return AuthResponse{
		User: UserResponse{
			ID:        result.User.ID,
			Email:     result.User.Email,
			FullName:  result.User.FullName,
			Status:    result.User.Status,
			CreatedAt: result.User.CreatedAt,
		},
		Tokens: TokenResponse{
			AccessToken:           result.Tokens.AccessToken,
			AccessTokenExpiresAt:  result.Tokens.AccessTokenExpiresAt,
			RefreshToken:          result.Tokens.RefreshToken,
			RefreshTokenExpiresAt: result.Tokens.RefreshTokenExpiresAt,
		},
	}
}
