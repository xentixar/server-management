package models

import (
	"context"
	"errors"

	"api/app/constants"
	"api/app/helpers"
	"api/config"

	"gorm.io/gorm"
)

type Organization struct {
	gorm.Model
	Name   string
	Slug   string `gorm:"uniqueIndex"`
	UserID uint

	User          User                       `gorm:"foreignKey:UserID"`
	Members       []OrganizationMember       `gorm:"foreignKey:OrganizationID"`
	MemberInvites []OrganizationMemberInvite `gorm:"foreignKey:OrganizationID"`
}

func (Organization) SlugAvailable(ctx context.Context, slug string) bool {
	_, err := gorm.G[Organization](config.GetDatabase()).Where("slug = ?", slug).First(ctx)
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func (Organization) GenerateUniqueSlug(ctx context.Context, name string) (string, error) {
	base := helpers.Slugify(name)

	for {
		suffix, err := helpers.RandomAlphanumeric(5)
		if err != nil {
			return "", err
		}

		slug := base + "-" + suffix
		if (Organization{}).SlugAvailable(ctx, slug) {
			return slug, nil
		}
	}
}

func (Organization) FindByID(ctx context.Context, id uint, withRelations ...string) (Organization, error) {
	query := gorm.
		G[Organization](config.GetDatabase()).
		Where("id = ?", id)

	for _, relation := range withRelations {
		query = query.Preload(relation, nil)
	}

	return query.First(ctx)
}

func (Organization) FindOwnedByID(ctx context.Context, id uint, ownerID uint, withRelations ...string) (Organization, error) {
	query := gorm.
		G[Organization](config.GetDatabase()).
		Where("id = ? AND user_id = ?", id, ownerID)

	for _, relation := range withRelations {
		query = query.Preload(relation, nil)
	}

	return query.First(ctx)
}

func (Organization) FindOwnedByUser(ctx context.Context, ownerID uint) ([]Organization, error) {
	return gorm.
		G[Organization](config.GetDatabase()).
		Where("user_id = ?", ownerID).
		Find(ctx)
}

func (o *Organization) Create(ctx context.Context) error {
	result := gorm.WithResult()
	err := gorm.G[Organization](config.GetDatabase(), result).Create(ctx, o)
	if err != nil {
		return err
	}

	return OrganizationMember{}.Upsert(ctx, o.ID, o.UserID, constants.Roles[constants.OwnerRole])
}

func (o *Organization) Update(ctx context.Context) error {
	_, err := gorm.G[Organization](config.GetDatabase()).Where("id = ?", o.ID).Updates(ctx, *o)
	return err
}

func (o *Organization) Delete(ctx context.Context) error {
	_, err := gorm.G[Organization](config.GetDatabase()).Where("id = ?", o.ID).Delete(ctx)
	return err
}
