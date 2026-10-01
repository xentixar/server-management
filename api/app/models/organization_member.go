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

func (OrganizationMember) Upsert(
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

	return gorm.G[OrganizationMember](config.GetDatabase(), clause.OnConflict{
		Columns: []clause.Column{
			{Name: "organization_id"},
			{Name: "user_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"role",
			"updated_at",
		}),
	}).Create(ctx, &member)
}

func (OrganizationMember) FindByID(
	ctx context.Context,
	id uint,
	organizationID uint,
	withRelations ...string,
) (OrganizationMember, error) {
	query := gorm.G[OrganizationMember](config.GetDatabase()).
		Where("id = ? AND organization_id = ?", id, organizationID)

	for _, relation := range withRelations {
		query = query.Preload(relation, nil)
	}

	return query.First(ctx)
}

func (OrganizationMember) FindByEmail(
	ctx context.Context,
	organizationID uint,
	email string,
) (*OrganizationMember, error) {
	member, err := gorm.G[OrganizationMember](config.GetDatabase()).
		Joins(clause.Has("User"), func(db gorm.JoinBuilder, joinTable clause.Table, curTable clause.Table) error {
			db.Where("?.email = ?", joinTable, email)
			return nil
		}).
		Where("organization_id = ?", organizationID).
		First(ctx)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (m *OrganizationMember) UpdateRole(ctx context.Context, role constants.Role) error {
	if !constants.IsValidRole(role.Name) || role.Name == constants.OwnerRole {
		return errors.New("invalid role")
	}
	if m.Role == constants.OwnerRole {
		return errors.New("cannot change owner role")
	}

	_, err := gorm.G[OrganizationMember](config.GetDatabase()).
		Where("id = ?", m.ID).
		Update(ctx, "role", role.Name)
	if err != nil {
		return err
	}

	m.Role = role.Name
	return nil
}

func (m *OrganizationMember) Delete(ctx context.Context) error {
	if m.Role == constants.OwnerRole {
		return errors.New("cannot remove owner")
	}

	_, err := gorm.G[OrganizationMember](config.GetDatabase()).
		Where("id = ?", m.ID).
		Delete(ctx)
	return err
}
