package models

import "gorm.io/gorm"

type PersonalAccessToken struct {
	gorm.Model
	UserID int
	Token  string `gorm:"uniqueIndex"`

	User User
}
