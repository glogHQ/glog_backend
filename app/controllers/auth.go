package controllers

import (
	"backend/app/auth"
	"backend/app/handlers"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type AuthController struct {
	authHandler *handlers.AuthHandler
}

func NewAuthController(authHandler *handlers.AuthHandler) *AuthController {
	return &AuthController{
		authHandler: authHandler,
	}
}

func (a *AuthController) Init(router *mux.Router) {
	router.HandleFunc("/login", a.login).Methods("POST", "OPTIONS")
	router.HandleFunc("/register", a.register).Methods("POST")
	router.HandleFunc("/refresh-token", a.refreshToken).Methods("POST")

}

func (a *AuthController) login(w http.ResponseWriter, r *http.Request) {
	loginRequest := &handlers.LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(loginRequest); err != nil {
		fmt.Println(err)
		return
	}
	loginResponse, loginErr := a.authHandler.Login(loginRequest)
	if loginErr != nil {
		fmt.Println(loginErr)
		return
	}

	http.SetCookie(w, loginResponse.Cookies.AuthTokenCookie)
	http.SetCookie(w, loginResponse.Cookies.RefreshTokenCookie)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loginResponse.User)
}

func (a *AuthController) register(w http.ResponseWriter, r *http.Request) {
	registerRequest := &handlers.RegisterRequest{}
	if err := json.NewDecoder(r.Body).Decode(registerRequest); err != nil {
		fmt.Println(err)
		return
	}
	registerResponse, registerErr := a.authHandler.Register(registerRequest)
	if registerErr != nil {
		fmt.Println(registerErr)
		return
	}

	http.SetCookie(w, registerResponse.Cookies.AuthTokenCookie)
	http.SetCookie(w, registerResponse.Cookies.RefreshTokenCookie)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(registerResponse.User)
}

func (a *AuthController) refreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.RefreshTokenCookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	authCookies, loginErr := a.authHandler.RefreshToken(cookie.Value)
	if loginErr != nil {
		fmt.Println(loginErr)
		return
	}

	http.SetCookie(w, authCookies.AuthTokenCookie)
	http.SetCookie(w, authCookies.RefreshTokenCookie)
	w.WriteHeader(http.StatusOK)
}
