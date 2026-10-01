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

	User          User                       `gorm:"foreignKey:UserID"`
	Members       []OrganizationMember       `gorm:"foreignKey:OrganizationID"`
	MemberInvites []OrganizationMemberInvite `gorm:"foreignKey:OrganizationID"`
}

func (Organization) IsSlugAvailable(ctx context.Context, slug string) bool {
	_, err := gorm.G[Organization](config.GetDatabase()).Where("slug = ?", slug).First(ctx)
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func (Organization) UniqueSlug(ctx context.Context, name string) (string, error) {
	base := utils.Slugify(name)

	for {
		suffix, err := utils.RandomAlphanumeric(5)
		if err != nil {
			return "", err
		}

		slug := base + "-" + suffix
		if (Organization{}).IsSlugAvailable(ctx, slug) {
			return slug, nil
		}
	}
}

func (Organization) FindByID(ctx context.Context, id uint, userID uint, withRelations ...string) (Organization, error) {
	query := gorm.
		G[Organization](config.GetDatabase()).
		Where("id = ? AND user_id = ?", id, userID)

	for _, relation := range withRelations {
		query = query.Preload(relation, nil)
	}

	return query.First(ctx)
}

func (Organization) FindByUserID(ctx context.Context, userID uint) ([]Organization, error) {
	return gorm.
		G[Organization](config.GetDatabase()).
		Where("user_id = ?", userID).
		Find(ctx)
}

func (o *Organization) Create(ctx context.Context) error {
	result := gorm.WithResult()
	err := gorm.G[Organization](config.GetDatabase(), result).Create(ctx, o)
	if err != nil {
		return err
	}

	return OrganizationMember{}.Assign(ctx, o.ID, o.UserID, constants.ROLES[constants.OwnerRole])
}

func (o *Organization) Update(ctx context.Context) error {
	_, err := gorm.G[Organization](config.GetDatabase()).Where("id = ?", o.ID).Updates(ctx, *o)
	return err
}

func (o *Organization) Delete(ctx context.Context) error {
	_, err := gorm.G[Organization](config.GetDatabase()).Where("id = ?", o.ID).Delete(ctx)
	return err
}
