package main

import (
	"backend/app/auth"
	"backend/app/handlers"
	"backend/app/models"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/harranali/authority"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Router      *mux.Router
	DB          *gorm.DB
	Authority   *authority.Authority
	UserAuth    *auth.UserAuth
	TokenAuth   *auth.TokenAuth
	AuthHandler *handlers.AuthHandler
}

func (a *App) Initialize() {
	db, err := gorm.Open(sqlite.Open("../test.db"))

	if err != nil {
		log.Fatal(err)
	}
	a.DB = db
	a.Router = mux.NewRouter()

	a.Migrate()
	a.InitializeAuthority()
	a.InitializeAuth()
	a.InitializeRoutes()
}

func (a *App) Migrate() {
	a.DB.AutoMigrate(&models.Post{})
}

func (a *App) InitializeAuthority() {
	a.Authority = authority.New(authority.Options{
		TablesPrefix: "authority_",
		DB:           a.DB,
	})
}

func (a *App) InitializeAuth() {
	a.UserAuth = auth.NewUserAuth(a.DB)
	a.TokenAuth = auth.NewTokenAuth(a.DB)
	a.AuthHandler = handlers.NewAuthHandler(a.UserAuth, a.TokenAuth, validate)
}

func (a *App) InitializeRoutes() {
	authWrapper := func(handler func(http.ResponseWriter, *http.Request)) http.Handler {
		return a.TokenAuth.AuthTokenMiddleware(http.HandlerFunc(handler))
	}
	a.Router.HandleFunc("/posts", a.getPosts).Methods("GET")
	a.Router.Handle("/posts", authWrapper(a.createPost)).Methods("POST")
	a.Router.HandleFunc("/posts/{id:[0-9]+}", a.getPost).Methods("GET")
	a.Router.Handle("/posts/{id:[0-9]+}", authWrapper(a.updatePost)).Methods("PATCH")
	a.Router.Handle("/posts/{id:[0-9]+}", authWrapper(a.deletePost)).Methods("DELETE")
	a.Router.HandleFunc("/auth/login", a.login).Methods("POST")
	a.Router.HandleFunc("/auth/register", a.register).Methods("POST")
	a.Router.HandleFunc("/auth/refresh-token", a.refreshToken).Methods("POST")
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func (a *App) Run() {
	log.Fatal(http.ListenAndServe(":8010", logRequest(a.Router)))
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	loginRequest := &handlers.LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(loginRequest); err != nil {
		fmt.Println(err)
		return
	}
	authCookies, loginErr := a.AuthHandler.Login(loginRequest)
	if loginErr != nil {
		fmt.Println(loginErr)
		return
	}

	http.SetCookie(w, authCookies.AuthTokenCookie)
	http.SetCookie(w, authCookies.RefreshTokenCookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (a *App) register(w http.ResponseWriter, r *http.Request) {
	registerRequest := &handlers.RegisterRequest{}
	if err := json.NewDecoder(r.Body).Decode(registerRequest); err != nil {
		fmt.Println(err)
		return
	}
	registerResponse, registerErr := a.AuthHandler.Register(registerRequest)
	if registerErr != nil {
		fmt.Println(registerErr)
		return
	}

	http.SetCookie(w, registerResponse.Cookies.AuthTokenCookie)
	http.SetCookie(w, registerResponse.Cookies.RefreshTokenCookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(registerResponse.User)
}

func (a *App) refreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.RefreshTokenCookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	authCookies, loginErr := a.AuthHandler.RefreshToken(cookie.Value)
	if loginErr != nil {
		fmt.Println(loginErr)
		return
	}

	http.SetCookie(w, authCookies.AuthTokenCookie)
	http.SetCookie(w, authCookies.RefreshTokenCookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (a *App) getPosts(w http.ResponseWriter, r *http.Request) {
	var posts []models.Post
	a.DB.Find(&posts)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

type createPostRequest struct {
	Title string `json:"title" validate:"required"`
	Body  string `json:"body" validate:"required"`
}

func (a *App) createPost(w http.ResponseWriter, r *http.Request) {
	input := &createPostRequest{}
	if err := json.NewDecoder(r.Body).Decode(input); err != nil {
		fmt.Println(err)
		return
	}

	err := validate.Struct(input)
	validationErrors := err.(validator.ValidationErrors)
	fmt.Println(validationErrors)

	user := r.Context().Value(auth.ContextUserKey).(*models.User)
	newPost := &models.Post{Title: input.Title, Body: input.Body, UserID: user.ID}

	a.DB.Create(newPost)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPost)
}

func (a *App) updatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIdString := vars["id"]
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		fmt.Println("Error converting id")
		return
	}

	input := &createPostRequest{}
	if err := json.NewDecoder(r.Body).Decode(input); err != nil {
		fmt.Println(err)
		return
	}

	valErr := validate.Struct(input)
	validationErrors := valErr.(validator.ValidationErrors)
	fmt.Println(validationErrors)

	updatePost := &models.Post{Model: gorm.Model{ID: uint(postId)}}
	updateData := models.Post{Title: input.Title, Body: input.Body}

	a.DB.Model(updatePost).Updates(updateData)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatePost)
}

func (a *App) getPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIdString := vars["id"]
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		fmt.Println("Error converting id")
		return
	}

	post := &models.Post{}
	dbErr := a.DB.First(post, postId).Error
	fmt.Println(dbErr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

func (a *App) deletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIdString := vars["id"]
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		fmt.Println("Error converting id")
		return
	}

	result := a.DB.Delete(&models.Post{}, postId)
	fmt.Println(result.RowsAffected)
	fmt.Println(result.Error)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

var validate *validator.Validate

func main() {
	fmt.Println("Starting")

	validate = validator.New()

	a := App{}
	a.Initialize()
	a.Run()
	sqlDB, err := a.DB.DB()
	if err == nil {
		sqlDB.Close()
	}
}
