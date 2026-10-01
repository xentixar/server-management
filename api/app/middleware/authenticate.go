// Package middleware provides Gin middleware for authentication and request logging.
package middleware

import (
	"net/http"
	"strings"

	"api/app/models"

	"github.com/gin-gonic/gin"
)

func RequireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader("Authorization")
		if authorizationHeader == "" {
			abortUnauthorized(ctx, "Authorization header missing")
			return
		}

		parts := strings.SplitN(authorizationHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
			abortUnauthorized(ctx, "Invalid authorization header")
			return
		}

		personalAccessToken, err := (models.PersonalAccessToken{}).FindByToken(ctx, parts[1])
		if err != nil {
			abortUnauthorized(ctx, "Invalid access token")
			return
		}

		ctx.Set("user", personalAccessToken.User)
		ctx.Next()
	}
}

func abortUnauthorized(ctx *gin.Context, message string) {
	ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"message": message,
		"data":    gin.H{},
		"success": false,
	})
}
