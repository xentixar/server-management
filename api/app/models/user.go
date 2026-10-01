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

	Organizations []Organization       `gorm:"foreignKey:UserID"`
	Memberships   []OrganizationMember `gorm:"foreignKey:UserID"`
}

func (User) FindByEmail(ctx context.Context, email string) (User, error) {
	return gorm.G[User](config.GetDatabase()).Where("email = ?", email).First(ctx)
}

func (User) IsEmailTaken(ctx context.Context, email string) bool {
	_, err := (User{}).FindByEmail(ctx, email)
	return err == nil
}

func (u *User) Create(ctx context.Context) error {
	return gorm.G[User](config.GetDatabase()).Create(ctx, u)
}
