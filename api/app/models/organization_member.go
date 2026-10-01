package models

import (
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
	organizationID uint,
	userID uint,
	role constants.Role,
) error {
	if !constants.IsValidRole(role) {
		return errors.New("invalid role")
	}

	member := OrganizationMember{
		UserID:         userID,
		OrganizationID: organizationID,
		Role:           role.Name,
	}

	return config.GetDatabase().
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
