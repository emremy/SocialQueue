package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	store        *Store
	tokenManager *TokenManager
}

func NewService(store *Store, tokenManager *TokenManager) *Service {
	return &Service{
		store:        store,
		tokenManager: tokenManager,
	}
}

func (s *Service) Register(
	ctx context.Context,
	email string,
	password string,
	fullName *string,
	userAgent *string,
	ipAddress *string,
) (*AuthResult, error) {
	email = normalizeEmail(email)

	existingUser, err := s.store.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Status:       "active",
	}

	if err := s.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	tokens, err := s.createSessionTokens(ctx, user.ID, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:   *user,
		Tokens: *tokens,
	}, nil
}

func (s *Service) Login(
	ctx context.Context,
	email string,
	password string,
	userAgent *string,
	ipAddress *string,
) (*AuthResult, error) {
	email = normalizeEmail(email)

	user, err := s.store.GetUserByEmail(ctx, email)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	if user.Status != "active" {
		return nil, ErrInactiveUser
	}

	if !CheckPassword(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	tokens, err := s.createSessionTokens(ctx, user.ID, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:   *user,
		Tokens: *tokens,
	}, nil
}

func (s *Service) Refresh(
	ctx context.Context,
	refreshToken string,
	userAgent *string,
	ipAddress *string,
) (*AuthResult, error) {
	refreshTokenHash := HashRefreshToken(refreshToken)

	session, err := s.store.GetValidSessionByRefreshTokenHash(ctx, refreshTokenHash)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	user, err := s.store.GetUserByID(ctx, session.UserID)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	if user.Status != "active" {
		return nil, ErrInactiveUser
	}

	if err := s.store.RevokeSession(ctx, refreshTokenHash); err != nil {
		return nil, err
	}

	tokens, err := s.createSessionTokens(ctx, user.ID, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:   *user,
		Tokens: *tokens,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	refreshTokenHash := HashRefreshToken(refreshToken)

	err := s.store.RevokeSession(ctx, refreshTokenHash)
	if errors.Is(err, ErrNotFound) {
		return nil
	}

	return err
}

func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (*User, error) {
	user, err := s.store.GetUserByID(ctx, userID)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	if user.Status != "active" {
		return nil, ErrInactiveUser
	}

	return user, nil
}

func (s *Service) createSessionTokens(
	ctx context.Context,
	userID uuid.UUID,
	userAgent *string,
	ipAddress *string,
) (*AuthTokens, error) {
	accessToken, accessTokenExpiresAt, err := s.tokenManager.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshTokenHash, refreshTokenExpiresAt, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	session := &UserSession{
		ID:               uuid.New(),
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		UserAgent:        userAgent,
		IPAddress:        ipAddress,
		ExpiresAt:        refreshTokenExpiresAt,
	}

	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
