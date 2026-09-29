// Package routes register all routes
package routes

import (
	"api/app/controllers"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.RouterGroup) {
	healthController := controllers.HealthController{}
	authController := controllers.AuthController{}

	router.GET("/health", healthController.GetHealth)
	group := router.Group("/auth")
	{
		group.POST("/login", authController.Login)
	}
}
