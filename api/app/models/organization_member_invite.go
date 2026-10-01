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
	if !constants.IsValidRole(role.Name) || role.Name == constants.OwnerRole {
		return errors.New("invalid role")
	}

	invite := OrganizationMemberInvite{
		Email:          email,
		OrganizationID: organizationID,
		Role:           role.Name,
	}

	return gorm.G[OrganizationMemberInvite](config.GetDatabase()).Create(ctx, &invite)
}

func (OrganizationMemberInvite) FindByID(
	ctx context.Context,
	id uint,
	organizationID uint,
) (OrganizationMemberInvite, error) {
	return gorm.G[OrganizationMemberInvite](config.GetDatabase()).
		Where("id = ? AND organization_id = ?", id, organizationID).
		First(ctx)
}

func (OrganizationMemberInvite) FindByEmail(
	ctx context.Context,
	organizationID uint,
	email string,
) (*OrganizationMemberInvite, error) {
	invite, err := gorm.G[OrganizationMemberInvite](config.GetDatabase()).
		Where("organization_id = ? AND email = ?", organizationID, email).
		First(ctx)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &invite, nil
}

func (OrganizationMemberInvite) FindByOrganization(
	ctx context.Context,
	organizationID uint,
) ([]OrganizationMemberInvite, error) {
	return gorm.G[OrganizationMemberInvite](config.GetDatabase()).
		Where("organization_id = ?", organizationID).
		Find(ctx)
}

func (i *OrganizationMemberInvite) Delete(ctx context.Context) error {
	_, err := gorm.G[OrganizationMemberInvite](config.GetDatabase()).
		Where("id = ?", i.ID).
		Delete(ctx)
	return err
}
