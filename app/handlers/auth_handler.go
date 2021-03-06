package handlers

import (
	"backend/app/auth"
	"backend/app/models"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type AuthCookies struct {
	AuthTokenCookie    *http.Cookie
	RefreshTokenCookie *http.Cookie
}

type LoginResponse struct {
	User    *models.User
	Cookies *AuthCookies
}

type RegisterRequest struct {
	Email            string `json:"email" validate:"required"`
	Password         string `json:"password" validate:"required,eqfield=PasswordRepeated"`
	PasswordRepeated string `json:"password_repeated" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthHandler struct {
	UserAuth  *auth.UserAuth
	TokenAuth *auth.TokenAuth
	Validator *validator.Validate
}

func NewAuthHandler(userAuth *auth.UserAuth, tokenAuth *auth.TokenAuth, validator *validator.Validate) *AuthHandler {
	return &AuthHandler{
		UserAuth:  userAuth,
		TokenAuth: tokenAuth,
		Validator: validator,
	}
}

func (a *AuthHandler) createAuthCookiesForUser(user *models.User) (*AuthCookies, error) {
	authToken, authTokenErr := a.TokenAuth.CreateAuthToken(user)
	if authTokenErr != nil {
		return nil, authTokenErr
	}
	authTokenCookie, refreshTokenCookie := a.TokenAuth.CreateAuthCookies(authToken)
	return &AuthCookies{
		AuthTokenCookie:    authTokenCookie,
		RefreshTokenCookie: refreshTokenCookie,
	}, nil
}

func (a *AuthHandler) Register(registerRequest *RegisterRequest) (*LoginResponse, error) {
	if valErr := a.Validator.Struct(registerRequest); valErr != nil {
		return nil, valErr.(validator.ValidationErrors)
	}
	user, userErr := a.UserAuth.CreateUser(registerRequest.Email, registerRequest.Password)
	if userErr != nil {
		return nil, userErr
	}

	authCookies, authCookiesErr := a.createAuthCookiesForUser(user)
	if authCookiesErr != nil {
		return nil, authCookiesErr
	}

	return &LoginResponse{
		User:    user,
		Cookies: authCookies,
	}, nil
}

func (a *AuthHandler) Login(loginRequest *LoginRequest) (*LoginResponse, error) {
	if valErr := a.Validator.Struct(loginRequest); valErr != nil {
		return nil, valErr.(validator.ValidationErrors)
	}
	user, checkUserErr := a.UserAuth.CheckUserPassword(loginRequest.Email, loginRequest.Password)
	if checkUserErr != nil {
		return nil, checkUserErr
	}

	authCookies, authCookiesErr := a.createAuthCookiesForUser(user)
	if authCookiesErr != nil {
		return nil, authCookiesErr
	}

	return &LoginResponse{
		User:    user,
		Cookies: authCookies,
	}, nil
}

func (a *AuthHandler) RefreshToken(refreshToken string) (*AuthCookies, error) {
	authToken, authTokenErr := a.TokenAuth.RefreshToken(refreshToken)
	if authTokenErr != nil {
		return nil, authTokenErr
	}
	authTokenCookie, refreshTokenCookie := a.TokenAuth.CreateAuthCookies(authToken)
	return &AuthCookies{
		AuthTokenCookie:    authTokenCookie,
		RefreshTokenCookie: refreshTokenCookie,
	}, nil
}
