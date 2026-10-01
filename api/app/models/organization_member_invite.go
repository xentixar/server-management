package models

import (
	"context"
	"errors"

	"api/app/constants"
	"api/config"

	"gorm.io/gorm"
)

type OrganizationMemberInvite struct {
	gorm.Model
	OrganizationID uint   `gorm:"uniqueIndex:idx_organization_email"`
	Email          string `gorm:"uniqueIndex:idx_organization_email"`
	Role           string `gorm:"type:varchar(50);not null"`

	Organization Organization `gorm:"foreignKey:OrganizationID"`
}

func (OrganizationMemberInvite) Create(
	ctx context.Context,
	organizationID uint,
	email string,
	role constants.Role,
) error {
	if !constants.IsValidRole(role.Name) {
		return errors.New("invalid role")
	}

	member := OrganizationMemberInvite{
		Email:          email,
		OrganizationID: organizationID,
		Role:           role.Name,
	}

	return gorm.G[OrganizationMemberInvite](config.GetDatabase()).Create(ctx, &member)
}

func (OrganizationMemberInvite) FindOneByEmail(
	ctx context.Context,
	organizationID uint,
	email string,
) (*OrganizationMemberInvite, error) {
	invite, err := gorm.G[OrganizationMemberInvite](config.GetDatabase()).
		Where("organization_id = ? and email = ?", organizationID, email).
		First(ctx)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &invite, nil
}
