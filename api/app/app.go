// Package app boots the HTTP server, wires dependencies, and runs the API.
package app

import (
	"errors"
	"log"
	"net/http"
	"time"

	"api/app/models"
	"api/app/routes"
	"api/app/validators"
	"api/config"
)

func Initialize() {
	config.LoadEnv()
	InitializeRouting()
	config.InitializeDatabase()
	validators.Register()

	group := GetRouter().Group("/api/v1")
	{
		routes.Register(group)
	}

	_ = config.GetDatabase().AutoMigrate(
		&models.User{},
		&models.PersonalAccessToken{},
		&models.Organization{},
		&models.OrganizationMember{},
		&models.OrganizationMemberInvite{},
		&models.Server{},
		&models.AgentJob{},
		&models.ServerService{},
		&models.ServerCron{},
	)
	_ = (models.AgentJob{}).BackfillUIDs()
}

func Run() {
	server := &http.Server{
		Addr:              ":" + config.MustGet("APP_PORT"),
		Handler:           GetRouter(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Listen error: %s\n", err)
	}
}
