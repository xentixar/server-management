// Package routes register all routes
package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.RouterGroup) {
	router.GET("/ping", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
}
