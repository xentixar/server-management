// Package config loads environment settings and exposes shared infrastructure such as the database.
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	AppPort     string
	Environment string
	DBHost      string
	DBPort      string
	DBUsername  string
	DBPassword  string
	DBDatabase  string
}

var env Env

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	env.AppPort = ensureKeyExists("APP_PORT")
	env.Environment = ensureKeyExists("APP_ENV")
	env.DBHost = ensureKeyExists("DB_HOST")
	env.DBPort = ensureKeyExists("DB_PORT")
	env.DBUsername = ensureKeyExists("DB_USERNAME")
	env.DBPassword = ensureKeyExists("DB_PASSWORD")
	env.DBDatabase = ensureKeyExists("DB_DATABASE")
}

func ENV(key string) string {
	switch key {
	case "APP_PORT":
		return env.AppPort
	case "APP_ENV":
		return env.Environment
	case "DB_HOST":
		return env.DBHost
	case "DB_PORT":
		return env.DBPort
	case "DB_USERNAME":
		return env.DBUsername
	case "DB_PASSWORD":
		return env.DBPassword
	case "DB_DATABASE":
		return env.DBDatabase
	default:
		panic("Unknown environment key: " + key)
	}
}

func ensureKeyExists(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Key " + key + " is missing from .env")
	}

	return value
}
