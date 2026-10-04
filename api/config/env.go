// Package config loads environment settings and exposes shared infrastructure such as the database.
package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Env struct {
	AppPort                    string
	Environment                string
	DBHost                     string
	DBPort                     string
	DBUsername                 string
	DBPassword                 string
	DBDatabase                 string
	PublicAPIURL               string
	PublicDashboardURL         string
	AppEncryptionKey           string
	AgentBinDir                string
	GitHubOAuthClientID        string
	GitHubOAuthClientSecret    string
	GitLabOAuthClientID        string
	GitLabOAuthClientSecret    string
	BitbucketOAuthClientID     string
	BitbucketOAuthClientSecret string
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
	env.PublicAPIURL = os.Getenv("PUBLIC_API_URL")
	env.PublicDashboardURL = os.Getenv("PUBLIC_DASHBOARD_URL")
	env.AppEncryptionKey = os.Getenv("APP_ENCRYPTION_KEY")
	env.AgentBinDir = os.Getenv("AGENT_BIN_DIR")
	if env.AgentBinDir == "" {
		env.AgentBinDir = "bin"
	}
	env.GitHubOAuthClientID = os.Getenv("GITHUB_OAUTH_CLIENT_ID")
	env.GitHubOAuthClientSecret = os.Getenv("GITHUB_OAUTH_CLIENT_SECRET")
	env.GitLabOAuthClientID = os.Getenv("GITLAB_OAUTH_CLIENT_ID")
	env.GitLabOAuthClientSecret = os.Getenv("GITLAB_OAUTH_CLIENT_SECRET")
	env.BitbucketOAuthClientID = os.Getenv("BITBUCKET_OAUTH_CLIENT_ID")
	env.BitbucketOAuthClientSecret = os.Getenv("BITBUCKET_OAUTH_CLIENT_SECRET")
}

func MustGet(key string) string {
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
	case "PUBLIC_API_URL":
		return env.PublicAPIURL
	default:
		panic("Unknown environment key: " + key)
	}
}

func PublicAPIURL() string {
	return env.PublicAPIURL
}

func PublicDashboardURL() string {
	return env.PublicDashboardURL
}

func AppEncryptionKey() string {
	return env.AppEncryptionKey
}

func GitHubOAuthClientID() string {
	return env.GitHubOAuthClientID
}

func GitHubOAuthClientSecret() string {
	return env.GitHubOAuthClientSecret
}

func GitLabOAuthClientID() string {
	return env.GitLabOAuthClientID
}

func GitLabOAuthClientSecret() string {
	return env.GitLabOAuthClientSecret
}

func BitbucketOAuthClientID() string {
	return env.BitbucketOAuthClientID
}

func BitbucketOAuthClientSecret() string {
	return env.BitbucketOAuthClientSecret
}

func AgentBinDir() string {
	if filepath.IsAbs(env.AgentBinDir) {
		return env.AgentBinDir
	}
	cwd, err := os.Getwd()
	if err != nil {
		return env.AgentBinDir
	}
	return filepath.Join(cwd, env.AgentBinDir)
}

func ensureKeyExists(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Key " + key + " is missing from .env")
	}

	return value
}
