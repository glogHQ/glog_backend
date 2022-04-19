package auth

import (
	"backend/app/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Auth struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *Auth {
	auth := &Auth{DB: db}
	auth.migrateTables()
	return auth
}

func (a *Auth) migrateTables() {
	a.DB.AutoMigrate(&models.User{})
}

func (a *Auth) CreateUser(email string, password string) (*models.User, error) {
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

func (a *Auth) CheckUserPassword(email string, password string) bool {
	user := &models.User{}
	result := a.DB.First(user, "email = ?", email)
	if result.Error != nil {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil
}
