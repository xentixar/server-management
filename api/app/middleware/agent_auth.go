package middleware

import (
	"strings"

	"api/app/models"

	"github.com/gin-gonic/gin"
)

func RequireAgent() gin.HandlerFunc {
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

		server, err := (models.Server{}).FindByAgentToken(ctx, parts[1])
		if err != nil {
			abortUnauthorized(ctx, "Invalid agent token")
			return
		}

		ctx.Set("server", server)
		ctx.Next()
	}
}
