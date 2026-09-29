package models

import (
	"time"

	"gorm.io/gorm"
)

type PersonalAccessToken struct {
	gorm.Model
	UserID    int
	Token     string `gorm:"uniqueIndex"`
	ExpiresAt *time.Time

	User User
}
