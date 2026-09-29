// Package routes register all routes
package routes

import (
	"api/app/controllers"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.RouterGroup) {
	router.GET("/health", controllers.HealthController{}.GetHealth)
}
