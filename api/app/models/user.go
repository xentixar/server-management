// Package models contains all models
package models

import (
	"context"

	"api/config"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string
	Email    string `gorm:"uniqueIndex"`
	Password string
}

func (User) FindByEmail(email string) (User, error) {
	return gorm.G[User](config.GetDatabase()).Where("email = ?", email).First(context.Background())
}

func (User) IsEmailTaken(email string) bool {
	_, err := (User{}).FindByEmail(email)
	return err == nil
}

func (u *User) Create() error {
	return gorm.G[User](config.GetDatabase()).Create(context.Background(), u)
}
