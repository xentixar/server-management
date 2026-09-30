package models

import (
	"context"
	"time"

	"api/config"

	"gorm.io/gorm"
)

type PersonalAccessToken struct {
	gorm.Model
	UserID    int
	Token     string `gorm:"uniqueIndex"`
	ExpiresAt *time.Time

	User User
}

func (PersonalAccessToken) FindByToken(token string) (PersonalAccessToken, error) {
	return gorm.
		G[PersonalAccessToken](config.GetDatabase()).
		Preload("User", nil).
		Where("token = ?", token).
		First(context.Background())
}

func (t *PersonalAccessToken) Create() error {
	return gorm.G[PersonalAccessToken](config.GetDatabase()).Create(context.Background(), t)
}
