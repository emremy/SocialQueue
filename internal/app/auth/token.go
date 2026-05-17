package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenManager struct {
	secret         []byte
	accessTokenTTL time.Duration
	refreshTTL     time.Duration
}

type AccessTokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func CreateTokenManager(secret string, accessTokenTTL time.Duration, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		secret:         []byte(secret),
		accessTokenTTL: accessTokenTTL,
		refreshTTL:     refreshTTL,
	}
}

func (m *TokenManager) GenerateAccessToken(userID uuid.UUID) (string, time.Time, error) {
	now := time.Now().UTC()
	expires := now.Add(m.accessTokenTTL)

	claims := AccessTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signedToken, expires, nil
}

func (m *TokenManager) VerifyAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}

		return m.secret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (m *TokenManager) GenerateRefreshToken() (plainToken string, tokenHash string, expiresAt time.Time, err error) {
	tokenBytes := make([]byte, 32)

	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", time.Time{}, err
	}

	plainToken = base64.RawURLEncoding.EncodeToString(tokenBytes)
	tokenHash = HashRefreshToken(plainToken)
	expiresAt = time.Now().UTC().Add(m.refreshTTL)

	return plainToken, tokenHash, expiresAt, nil
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
