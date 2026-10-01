package models

import (
	"context"
	"errors"

	"api/app/constants"
	"api/config"
	"api/utils"

	"gorm.io/gorm"
)

type Organization struct {
	gorm.Model
	Name   string
	Slug   string `gorm:"uniqueIndex"`
	UserID uint

	User    User                 `gorm:"foreignKey:UserID"`
	Members []OrganizationMember `gorm:"foreignKey:OrganizationID"`
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

func (Organization) FindByID(id uint, userID uint, withRelations ...string) (Organization, error) {
	query := gorm.
		G[Organization](config.GetDatabase()).
		Where("id = ? AND user_id = ?", id, userID)

	for _, relation := range withRelations {
		query = query.Preload(relation, nil)
	}

	return query.First(context.Background())
}

func (Organization) FindByUserID(userID uint) ([]Organization, error) {
	return gorm.
		G[Organization](config.GetDatabase()).
		Where("user_id = ?", userID).
		Find(context.Background())
}

func (o *Organization) Create() error {
	result := gorm.WithResult()
	err := gorm.G[Organization](config.GetDatabase(), result).Create(context.Background(), o)
	if err != nil {
		return err
	}

	return OrganizationMember{}.Assign(o.ID, o.UserID, constants.ROLES[constants.OwnerRole])
}

func (o *Organization) Update() error {
	_, err := gorm.G[Organization](config.GetDatabase()).Where("id = ?", o.ID).Updates(context.Background(), *o)
	return err
}

func (o *Organization) Delete() error {
	_, err := gorm.G[Organization](config.GetDatabase()).Where("id = ?", o.ID).Delete(context.Background())
	return err
}
