package main

import (
	"backend/app/auth"
	"backend/app/controllers"
	"backend/app/handlers"
	"backend/app/middlewares"
	"backend/app/models"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	mux_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/harranali/authority"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	a.InitializeHandlers()
	a.InitializeControllers()
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
}

func (a *App) InitializeHandlers() {
	validate := validator.New()
	a.AuthHandler = handlers.NewAuthHandler(a.UserAuth, a.TokenAuth, validate)
	a.PostsHandler = handlers.NewPostHandler(a.DB, validate)
}

func (a *App) InitializeControllers() {
	cors := mux_handlers.CORS(
		mux_handlers.AllowedHeaders([]string{"content-type"}),
		mux_handlers.AllowedOrigins([]string{"http://localhost:3000"}),
		mux_handlers.AllowCredentials(),
	)

	a.Router.Use(cors, middlewares.JSONResponseMiddleware)

	postsRouter := a.Router.PathPrefix("/posts").Subrouter()
	postsController := controllers.NewPostsController(a.TokenAuth, a.PostsHandler)
	postsController.Init(postsRouter)

	authRouter := a.Router.PathPrefix("/auth").Subrouter()
	authController := controllers.NewAuthController(a.AuthHandler)
	authController.Init(authRouter)

	usersRouter := a.Router.PathPrefix("/user").Subrouter()
	usersController := controllers.NewUsersController(a.TokenAuth)
	usersController.Init(usersRouter)
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

func main() {
	fmt.Println("Starting")

	a := App{}
	a.Initialize()
	a.Run()
	sqlDB, err := a.DB.DB()
	if err == nil {
		sqlDB.Close()
	}

	defer sqlDB.Close()
}
