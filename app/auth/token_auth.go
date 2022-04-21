package auth

import (
	"backend/app/models"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"gorm.io/gorm"
	"time"
)

var TokenExpiredError = errors.New("token is expired")

type TokenAuth struct {
	DB *gorm.DB
}

func NewTokenAuth(db *gorm.DB) *TokenAuth {
	auth := &TokenAuth{DB: db}
	auth.migrateTables()
	return auth
}

func (a *TokenAuth) migrateTables() {
	a.DB.AutoMigrate(&models.RefreshToken{})
	a.DB.AutoMigrate(&models.AuthToken{})
}

func (a *TokenAuth) CreateAuthToken(user *models.User) (*models.AuthToken, error) {
	randomAccessToken := make([]byte, 32)
	randomRefreshToken := make([]byte, 32)

	_, accessTokenErr := rand.Read(randomAccessToken)
	if accessTokenErr != nil {
		return nil, accessTokenErr
	}
	_, refreshTokenErr := rand.Read(randomRefreshToken)
	if refreshTokenErr != nil {
		return nil, accessTokenErr
	}

	authToken := &models.AuthToken{
		Token:     base64.URLEncoding.EncodeToString(randomAccessToken),
		ExpiresAt: time.Now().Add(time.Minute * 10),
		RefreshToken: models.RefreshToken{
			Token:     base64.URLEncoding.EncodeToString(randomRefreshToken),
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
		},
		UserID: user.ID,
	}
	result := a.DB.Create(authToken)
	if result.Error != nil {
		return nil, result.Error
	}

	return authToken, nil
}

func (a *TokenAuth) CheckAuthToken(token string) (*models.User, error) {
	authToken := &models.AuthToken{}
	result := a.DB.Preload("User").First(authToken, "token = ?", token)
	if result.Error != nil {
		return nil, result.Error
	}

	if authToken.ExpiresAt.Before(time.Now()) {
		return nil, TokenExpiredError
	}

	return &authToken.User, nil
}
