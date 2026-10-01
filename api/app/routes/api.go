// Package routes registers the API v1 HTTP endpoints.
package routes

import (
	"api/app/controllers"
	"api/app/middleware"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.RouterGroup) {
	health := controllers.HealthController{}
	auth := controllers.AuthController{}
	organizations := controllers.OrganizationController{}
	members := controllers.OrganizationMemberController{}
	invites := controllers.OrganizationMemberInviteController{}

	router.GET("/health", health.Check)

	group := router.Group("/auth")
	{
		group.POST("/login", auth.Login)
		group.POST("/register", auth.Register)
		group.GET("/user", middleware.RequireAuth(), auth.Me)
	}

	group = router.Group("/organizations", middleware.RequireAuth())
	{
		group.GET("", organizations.List)
		group.POST("", organizations.Create)
		group.GET("/:organization", organizations.Get)
		group.PATCH("/:organization", organizations.Update)
		group.DELETE("/:organization", organizations.Delete)

		group.GET("/:organization/members", members.List)
		group.PATCH("/:organization/members/:member", members.Update)
		group.DELETE("/:organization/members/:member", members.Delete)

		group.GET("/:organization/invites", invites.List)
		group.POST("/:organization/invites", invites.Create)
		group.POST("/:organization/invites/accept", invites.Accept)
		group.DELETE("/:organization/invites/:invite", invites.Delete)
	}
}
