package models

import (
	"gorm.io/gorm"
	"time"
)

type AuthToken struct {
	gorm.Model
	Token          string `gorm:"unique"`
	ExpiresAt      time.Time
	RefreshToken   RefreshToken
	RefreshTokenID uint
	User           User
	UserID         uint
}

type RefreshToken struct {
	gorm.Model
	Token     string `gorm:"unique"`
	ExpiresAt time.Time
}
