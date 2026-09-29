// Package config contains all configuration related stuffs
package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	err error
)

func InitializeDatabase() {
	dsn := "host=" + ENV("DB_HOST") +
		" user=" + ENV("DB_USERNAME") +
		" password=" + ENV("DB_PASSWORD") +
		" dbname=" + ENV("DB_DATABASE") +
		" port=" + ENV("DB_PORT") +
		" sslmode=disable"

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
}

func GetDatabase() *gorm.DB {
	return db
}
