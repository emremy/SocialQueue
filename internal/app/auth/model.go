package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:text;not null"`
	FullName     *string   `gorm:"type:varchar(150)"`
	Status       string    `gorm:"type:varchar(30);not null;default:active"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (User) TableName() string {
	return "users"
}

type UserSession struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	User   User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

	RefreshTokenHash string `gorm:"type:text;uniqueIndex;not null"`

	UserAgent *string `gorm:"type:text"`
	IPAddress *string `gorm:"type:inet"`

	ExpiresAt time.Time  `gorm:"not null;index"`
	RevokedAt *time.Time `gorm:"index"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserSession) TableName() string {
	return "user_sessions"
}

type AuthTokens struct {
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}

type AuthResult struct {
	User   User
	Tokens AuthTokens
}
