// Package routes registers the API v1 HTTP endpoints.
package routes

import (
	"api/app/controllers"
	"api/app/middleware"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.RouterGroup) {
	healthController := controllers.HealthController{}
	authController := controllers.AuthController{}
	organizationController := controllers.OrganizationController{}
	organizationMemberController := controllers.OrganizationMemberController{}
	organizationMemberInviteController := controllers.OrganizationMemberInviteController{}

	router.GET("/health", healthController.GetHealth)

	group := router.Group("/auth")
	{
		group.POST("/login", authController.Login)
		group.POST("/register", authController.Register)
		group.GET("/user", middleware.Authenticate(), authController.GetUser)
	}

	group = router.Group("/organizations", middleware.Authenticate())
	{
		group.GET("", organizationController.Index)
		group.POST("", organizationController.Store)
		group.GET("/:organization", organizationController.Show)
		group.PATCH("/:organization", organizationController.Update)
		group.DELETE("/:organization", organizationController.Destroy)

		group.GET("/:organization/members", organizationMemberController.Index)
		group.POST("/:organization/members/invite", organizationMemberInviteController.Invite)
	}
}
