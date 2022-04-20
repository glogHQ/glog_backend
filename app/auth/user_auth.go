package auth

import (
	"backend/app/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserAuth struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *UserAuth {
	auth := &UserAuth{DB: db}
	auth.migrateTables()
	return auth
}

func (a *UserAuth) migrateTables() {
	a.DB.AutoMigrate(&models.User{})
}

func (a *UserAuth) CreateUser(email string, password string) (*models.User, error) {
	hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(password), 8)
	if hashErr != nil {
		return nil, hashErr
	}

	user := &models.User{Email: email, Password: string(hashedPassword)}
	result := a.DB.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

func (a *UserAuth) CheckUserPassword(email string, password string) (*models.User, error) {
	user := &models.User{}
	result := a.DB.First(user, "email = ?", email)
	if result.Error != nil {
		return nil, result.Error
	}

	compareErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if compareErr != nil {
		return nil, compareErr
	}

	return user, nil
}
