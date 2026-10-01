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
	dsn := "host=" + MustGet("DB_HOST") +
		" user=" + MustGet("DB_USERNAME") +
		" password=" + MustGet("DB_PASSWORD") +
		" dbname=" + MustGet("DB_DATABASE") +
		" port=" + MustGet("DB_PORT") +
		" sslmode=disable"

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
}

func GetDatabase() *gorm.DB {
	return db
}
