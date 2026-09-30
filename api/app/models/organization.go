package models

import (
	"context"
	"errors"

	"api/config"
	"api/utils"

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

func (Organization) UniqueSlug(name string) (string, error) {
	base := utils.Slugify(name)

	for {
		suffix, err := utils.RandomAlphanumeric(5)
		if err != nil {
			return "", err
		}

		slug := base + "-" + suffix
		if (Organization{}).IsSlugAvailable(slug) {
			return slug, nil
		}
	}
}

func (Organization) FindByUserID(userID uint) ([]Organization, error) {
	return gorm.
		G[Organization](config.GetDatabase()).
		Where("user_id = ?", userID).
		Find(context.Background())
}

func (o *Organization) Create() error {
	return gorm.G[Organization](config.GetDatabase()).Create(context.Background(), o)
}
