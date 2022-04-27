package handlers

import (
	"backend/app/models"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type CreatePostRequest struct {
	Title string `json:"title" validate:"required"`
	Body  string `json:"body" validate:"required"`
}
type PostsHandler struct {
	DB        *gorm.DB
	Validator *validator.Validate
}

func NewPostHandler(db *gorm.DB, validator *validator.Validate) *PostsHandler {
	return &PostsHandler{
		DB:        db,
		Validator: validator,
	}
}
func (p *PostsHandler) GetPosts() *[]models.Post {
	var posts []models.Post
	p.DB.Find(&posts)
	return &posts
}

func (p *PostsHandler) GetPost(id uint) (*models.Post, error) {
	post := &models.Post{}
	if dbErr := p.DB.First(post, id).Error; dbErr != nil {
		return nil, dbErr
	}

	return post, nil
}

func (p *PostsHandler) DeletePost(id uint) error {
	return p.DB.Delete(&models.Post{}, id).Error
}

func (p *PostsHandler) CreatePost(request *CreatePostRequest, user *models.User) (*models.Post, error) {
	if valErr := p.Validator.Struct(request); valErr != nil {
		return nil, valErr.(validator.ValidationErrors)
	}
	newPost := &models.Post{Title: request.Title, Body: request.Body, UserID: user.ID}

	if dbErr := p.DB.Create(newPost).Error; dbErr != nil {
		return nil, dbErr
	}
	return newPost, nil
}

func (p *PostsHandler) UpdatePost(request *CreatePostRequest, id uint) (*models.Post, error) {
	if valErr := p.Validator.Struct(request); valErr != nil {
		return nil, valErr.(validator.ValidationErrors)
	}
	updatePost := &models.Post{Model: gorm.Model{ID: id}}
	updateData := models.Post{Title: request.Title, Body: request.Body}

	if dbErr := p.DB.Model(updatePost).Updates(updateData).Error; dbErr != nil {
		return nil, dbErr
	}

	return &updateData, nil
}
