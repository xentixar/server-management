// Package app contains application logics
package app

import (
	"errors"
	"log"
	"net/http"
	"time"

	"api/app/models"
	"api/config"
	"api/routes"
)

func Initialize() {
	config.LoadEnv()
	config.InitializeRouting()
	config.InitializeDatabase()

	group := config.GetRouter().Group("/api/v1")
	{
		routes.Register(group)
	}

	_ = config.GetDatabase().AutoMigrate(
		&models.User{},
		&models.PersonalAccessToken{},
		&models.Organization{},
		&models.OrganizationMember{},
	)
}

func Run() {
	server := &http.Server{
		Addr:              ":" + config.ENV("APP_PORT"),
		Handler:           config.GetRouter(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Listen error: %s\n", err)
	}
}
