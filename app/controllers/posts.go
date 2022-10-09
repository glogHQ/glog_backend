package controllers

import (
	"backend/app/auth"
	"backend/app/handlers"
	"backend/app/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type PostsController struct {
	tokenAuth    *auth.TokenAuth
	postsHandler *handlers.PostsHandler
}

func NewPostsController(tokenAuth *auth.TokenAuth, postsHandler *handlers.PostsHandler) *PostsController {
	return &PostsController{tokenAuth: tokenAuth, postsHandler: postsHandler}
}

func (p *PostsController) Init(router *mux.Router) {
	authWrapper := func(handler func(http.ResponseWriter, *http.Request)) http.Handler {
		return p.tokenAuth.AuthTokenMiddleware(http.HandlerFunc(handler))
	}

	router.HandleFunc("/posts", p.getPosts).Methods("GET")
	router.Handle("/posts", authWrapper(p.createPost)).Methods("POST")
	router.HandleFunc("/posts/{id:[0-9]+}", p.getPost).Methods("GET")
	router.Handle("/posts/{id:[0-9]+}", authWrapper(p.updatePost)).Methods("PATCH")
	router.Handle("/posts/{id:[0-9]+}", authWrapper(p.deletePost)).Methods("DELETE")
}
func (p *PostsController) getPosts(w http.ResponseWriter, r *http.Request) {
	posts := p.postsHandler.GetPosts()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

func (p *PostsController) createPost(w http.ResponseWriter, r *http.Request) {
	input := &handlers.CreatePostRequest{}
	if err := json.NewDecoder(r.Body).Decode(input); err != nil {
		fmt.Println(err)
		return
	}

	user := r.Context().Value(auth.ContextUserKey).(*models.User)
	post, createErr := p.postsHandler.CreatePost(input, user)
	if createErr != nil {
		fmt.Println(createErr)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func (p *PostsController) updatePost(w http.ResponseWriter, r *http.Request) {
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

	updatedPost, updateErr := p.postsHandler.UpdatePost(input, uint(postId))
	if updateErr != nil {
		fmt.Println(updateErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedPost)
}

func (p *PostsController) getPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIdString := vars["id"]
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		fmt.Println("Error converting id")
		return
	}

	post, postErr := p.postsHandler.GetPost(uint(postId))
	if postErr != nil {
		fmt.Println(postErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

func (p *PostsController) deletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postIdString := vars["id"]
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		fmt.Println("Error converting id")
		return
	}
	if deleteErr := p.postsHandler.DeletePost(uint(postId)); deleteErr != nil {
		fmt.Println(deleteErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
