package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) CreateUser(ctx context.Context, user *User) error {

	return s.db.WithContext(ctx).Create(user).Error
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User

	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User

	err := s.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Store) CreateSession(ctx context.Context, session *UserSession) error {
	return s.db.WithContext(ctx).Create(session).Error
}

func (s *Store) GetValidSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*UserSession, error) {
	var session UserSession

	err := s.db.WithContext(ctx).
		Where("refresh_token_hash = ?", refreshTokenHash).
		Where("expires_at > ?", time.Now().UTC()).
		Where("revoked_at IS NULL").
		First(&session).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *Store) RevokeSession(ctx context.Context, refreshTokenHash string) error {
	now := time.Now().UTC()

	result := s.db.WithContext(ctx).
		Model(&UserSession{}).
		Where("refresh_token_hash = ?", refreshTokenHash).
		Where("revoked_at IS NULL").
		Update("revoked_at", now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Store) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()

	return s.db.WithContext(ctx).
		Model(&UserSession{}).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Update("revoked_at", now).
		Error
}
