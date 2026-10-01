package app

import (
	"api/app/middleware"
	"api/config"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func InitializeRouting() {
	if config.MustGet("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router = gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
}

func GetRouter() *gin.Engine {
	return router
}
