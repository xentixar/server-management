package models

import (
	"context"
	"errors"

	"api/app/constants"
	"api/config"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrganizationMember struct {
	gorm.Model

	UserID         uint   `gorm:"not null;uniqueIndex:idx_org_member_user"`
	OrganizationID uint   `gorm:"not null;uniqueIndex:idx_org_member_user"`
	Role           string `gorm:"type:varchar(50);not null"`

	User         User         `gorm:"foreignKey:UserID"`
	Organization Organization `gorm:"foreignKey:OrganizationID"`
}

func (OrganizationMember) Assign(
	ctx context.Context,
	organizationID uint,
	userID uint,
	role constants.Role,
) error {
	if !constants.IsValidRole(role.Name) {
		return errors.New("invalid role")
	}

	member := OrganizationMember{
		UserID:         userID,
		OrganizationID: organizationID,
		Role:           role.Name,
	}

	return config.GetDatabase().
		WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "organization_id"},
				{Name: "user_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"role",
				"updated_at",
			}),
		}).
		Create(&member).Error
}

func (OrganizationMember) FindOneByEmail(
	ctx context.Context,
	organizationID uint,
	email string,
) (*OrganizationMember, error) {
	user, err := (User{}).FindByEmail(ctx, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	member, err := gorm.G[OrganizationMember](config.GetDatabase()).
		Where("organization_id = ? AND user_id = ?", organizationID, user.ID).
		First(ctx)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &member, nil
}
