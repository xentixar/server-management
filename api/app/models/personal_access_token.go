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

func (PersonalAccessToken) FindByToken(ctx context.Context, token string) (PersonalAccessToken, error) {
	return gorm.
		G[PersonalAccessToken](config.GetDatabase()).
		Preload("User", nil).
		Where("token = ?", token).
		First(ctx)
}

func (t *PersonalAccessToken) Create(ctx context.Context) error {
	return gorm.G[PersonalAccessToken](config.GetDatabase()).Create(ctx, t)
}
