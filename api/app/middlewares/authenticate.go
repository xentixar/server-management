// Package middlewares contains application level middlewares
package middlewares

import (
	"context"
	"net/http"
	"strings"

	"api/app/models"
	"api/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader("Authorization")

		if authorizationHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"message": "Authorization header missing",
				"data":    gin.H{},
				"success": false,
			})
			ctx.Abort()
			return
		}

		parts := strings.SplitN(authorizationHeader, " ", 2)

		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid authorization header",
				"data":    gin.H{},
				"success": false,
			})
			ctx.Abort()
			return
		}

		token := parts[1]

		bgContext := context.Background()

		personalAccessToken, err := gorm.G[models.PersonalAccessToken](config.GetDatabase()).Preload("User", nil).Where("token = ?", token).First(bgContext)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid access token",
				"data":    gin.H{},
				"success": false,
			})
			ctx.Abort()
			return
		}

		ctx.Set("user", personalAccessToken.User)

		ctx.Next()
	}
}
