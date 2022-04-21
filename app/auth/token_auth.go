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

	if _, accessTokenErr := rand.Read(randomAccessToken); accessTokenErr != nil {
		return nil, accessTokenErr
	}
	if _, refreshTokenErr := rand.Read(randomRefreshToken); refreshTokenErr != nil {
		return nil, refreshTokenErr
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

	if result := a.DB.Create(authToken); result.Error != nil {
		return nil, result.Error
	}

	return authToken, nil
}

func (a *TokenAuth) CheckAuthToken(token string) (*models.User, error) {
	authToken := &models.AuthToken{}
	if result := a.DB.Preload("User").First(authToken, "token = ?", token); result.Error != nil {
		return nil, result.Error
	}

	if authToken.ExpiresAt.Before(time.Now()) {
		return nil, TokenExpiredError
	}

	return &authToken.User, nil
}

func (a *TokenAuth) RefreshToken(token string) (*models.AuthToken, error) {
	refreshToken := &models.RefreshToken{}
	if result := a.DB.First(refreshToken, "token = ?", token); result.Error != nil {
		return nil, result.Error
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return nil, TokenExpiredError
	}

	authToken := &models.AuthToken{}
	if authTokenResult := a.DB.Preload("User").First(&authToken, refreshToken.ID); authTokenResult.Error != nil {
		return nil, authTokenResult.Error
	}

	return a.CreateAuthToken(&authToken.User)
}
