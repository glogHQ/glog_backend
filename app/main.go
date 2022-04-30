package main

import (
	"backend/app/auth"
	"backend/app/handlers"
	"backend/app/models"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	mux_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/harranali/authority"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Router       *mux.Router
	DB           *gorm.DB
	Authority    *authority.Authority
	UserAuth     *auth.UserAuth
	TokenAuth    *auth.TokenAuth
	AuthHandler  *handlers.AuthHandler
	PostsHandler *handlers.PostsHandler
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
	a.PostsHandler = handlers.NewPostHandler(a.DB, validate)
}

func (a *App) InitializeRoutes() {
	cors := mux_handlers.CORS(
		mux_handlers.AllowedHeaders([]string{"content-type"}),
		mux_handlers.AllowedOrigins([]string{"http://localhost:3000"}),
		mux_handlers.AllowCredentials(),
	)

	a.Router.Use(cors)

	authWrapper := func(handler func(http.ResponseWriter, *http.Request)) http.Handler {
		return a.TokenAuth.AuthTokenMiddleware(http.HandlerFunc(handler))
	}
	a.Router.HandleFunc("/posts", a.getPosts).Methods("GET")
	a.Router.Handle("/posts", authWrapper(a.createPost)).Methods("POST")
	a.Router.HandleFunc("/posts/{id:[0-9]+}", a.getPost).Methods("GET")
	a.Router.Handle("/posts/{id:[0-9]+}", authWrapper(a.updatePost)).Methods("PATCH")
	a.Router.Handle("/posts/{id:[0-9]+}", authWrapper(a.deletePost)).Methods("DELETE")
	a.Router.HandleFunc("/auth/login", a.login).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/auth/register", a.register).Methods("POST")
	a.Router.HandleFunc("/auth/refresh-token", a.refreshToken).Methods("POST")
	a.Router.Handle("/user", authWrapper(a.getUser)).Methods("GET", "OPTIONS")
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

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.ContextUserKey).(*models.User)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	loginRequest := &handlers.LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(loginRequest); err != nil {
		fmt.Println(err)
		return
	}
	loginResponse, loginErr := a.AuthHandler.Login(loginRequest)
	if loginErr != nil {
		fmt.Println(loginErr)
		return
	}

	http.SetCookie(w, loginResponse.Cookies.AuthTokenCookie)
	http.SetCookie(w, loginResponse.Cookies.RefreshTokenCookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loginResponse.User)
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
	posts := a.PostsHandler.GetPosts()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

func (a *App) createPost(w http.ResponseWriter, r *http.Request) {
	input := &handlers.CreatePostRequest{}
	if err := json.NewDecoder(r.Body).Decode(input); err != nil {
		fmt.Println(err)
		return
	}

	user := r.Context().Value(auth.ContextUserKey).(*models.User)
	post, createErr := a.PostsHandler.CreatePost(input, user)
	if createErr != nil {
		fmt.Println(createErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func (a *App) updatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIdString := vars["id"]
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		fmt.Println("Error converting id")
		return
	}

	input := &handlers.CreatePostRequest{}
	if err := json.NewDecoder(r.Body).Decode(input); err != nil {
		fmt.Println(err)
		return
	}

	updatedPost, updateErr := a.PostsHandler.UpdatePost(input, uint(postId))
	if updateErr != nil {
		fmt.Println(updateErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedPost)
}

func (a *App) getPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIdString := vars["id"]
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		fmt.Println("Error converting id")
		return
	}

	post, postErr := a.PostsHandler.GetPost(uint(postId))
	if postErr != nil {
		fmt.Println(postErr)
		return
	}

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
	if deleteErr := a.PostsHandler.DeletePost(uint(postId)); deleteErr != nil {
		fmt.Println(deleteErr)
		return
	}

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
