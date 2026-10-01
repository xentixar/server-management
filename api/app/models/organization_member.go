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

func (OrganizationMember) FindOneByEmail(
	ctx context.Context,
	organizationID uint,
	email string,
) (*OrganizationMember, error) {
	member, err := gorm.G[OrganizationMember](config.GetDatabase()).
		Joins(clause.Has("User"), func(db gorm.JoinBuilder, joinTable clause.Table, curTable clause.Table) error {
			db.Where("?.email = ?", clause.Column{Table: joinTable.Name, Name: "email"}, email)
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
