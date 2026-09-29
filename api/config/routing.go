package config

import "github.com/gin-gonic/gin"

var router *gin.Engine

func InitializeRouting() {
	if ENV("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router = gin.New()
	router.Use(gin.Recovery())
	router.Use(Logger())
}

func GetRouter() *gin.Engine {
	return router
}
