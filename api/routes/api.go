// Package routes register all routes
package routes

import (
	"api/app/controllers"
	"api/app/middlewares"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.RouterGroup) {
	healthController := controllers.HealthController{}
	authController := controllers.AuthController{}
	organizationController := controllers.OrganizationController{}

	router.GET("/health", healthController.GetHealth)
	group := router.Group("/auth")
	{
		group.POST("/login", authController.Login)
		group.POST("/register", authController.Register)
		group.GET("/user", middlewares.Authenticate(), authController.GetUser)
	}

	group = router.Group("/organizations")
	{
		group.GET("/", middlewares.Authenticate(), organizationController.Index)
	}
}
