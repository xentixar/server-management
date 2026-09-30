package models

import (
	"context"
	"errors"

	"api/config"

	"gorm.io/gorm"
)

type Organization struct {
	gorm.Model
	Name   string
	Slug   string `gorm:"uniqueIndex"`
	UserID uint

	User User
}

func (Organization) IsSlugAvailable(slug string) bool {
	_, err := gorm.G[Organization](config.GetDatabase()).Where("slug = ?", slug).First(context.Background())
	return errors.Is(err, gorm.ErrRecordNotFound)
}
