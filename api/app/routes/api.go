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
	servers := controllers.ServerController{}
	agent := controllers.AgentController{}

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

		group.GET("/:organization/servers", servers.List)
		group.POST("/:organization/servers/test-connection", servers.TestConnection)
		group.POST("/:organization/servers", servers.Create)
		group.GET("/:organization/servers/:server", servers.Get)
		group.GET("/:organization/servers/:server/metrics", servers.GetMetrics)
		group.POST("/:organization/servers/:server/bootstrap", servers.Bootstrap)
		group.POST("/:organization/servers/:server/provision", servers.Provision)
		group.GET("/:organization/servers/:server/provision-secrets", servers.GetProvisionSecrets)
		group.DELETE("/:organization/servers/:server/provision-secrets", servers.DismissProvisionSecrets)
		group.GET("/:organization/servers/:server/runtimes", servers.GetRuntimes)
		group.GET("/:organization/servers/:server/runtimes/catalog", servers.RuntimesCatalog)
		group.POST("/:organization/servers/:server/runtimes", servers.InstallRuntime)
		group.DELETE("/:organization/servers/:server/runtimes", servers.RemoveRuntime)
		group.PATCH("/:organization/servers/:server/runtimes/default", servers.SetRuntimeDefault)
		group.GET("/:organization/servers/:server/services", servers.ListServices)
		group.POST("/:organization/servers/:server/services", servers.CreateService)
		group.PATCH("/:organization/servers/:server/services/:service", servers.UpdateService)
		group.DELETE("/:organization/servers/:server/services/:service", servers.DeleteService)
		group.POST("/:organization/servers/:server/services/:service/control", servers.ControlService)
		group.GET("/:organization/servers/:server/crons", servers.ListCrons)
		group.POST("/:organization/servers/:server/crons", servers.CreateCron)
		group.PATCH("/:organization/servers/:server/crons/:cron", servers.UpdateCron)
		group.DELETE("/:organization/servers/:server/crons/:cron", servers.DeleteCron)
		group.POST("/:organization/servers/:server/crons/:cron/logs", servers.FetchCronLogs)
		group.GET("/:organization/servers/:server/firewall/rules", servers.ListFirewallRules)
		group.POST("/:organization/servers/:server/firewall/rules", servers.CreateFirewallRule)
		group.PATCH("/:organization/servers/:server/firewall/rules/:rule", servers.UpdateFirewallRule)
		group.DELETE("/:organization/servers/:server/firewall/rules/:rule", servers.DeleteFirewallRule)
		group.POST("/:organization/servers/:server/firewall/sync", servers.SyncFirewall)
		group.GET("/:organization/servers/:server/ssh-keys", servers.ListSshKeys)
		group.POST("/:organization/servers/:server/ssh-keys", servers.CreateSshKey)
		group.PATCH("/:organization/servers/:server/ssh-keys/:key", servers.UpdateSshKey)
		group.DELETE("/:organization/servers/:server/ssh-keys/:key", servers.DeleteSshKey)
		group.POST("/:organization/servers/:server/ssh-keys/sync", servers.SyncSshKeys)
		group.GET("/:organization/servers/:server/jobs/latest", servers.GetLatestJob)
		group.GET("/:organization/servers/:server/jobs/:job", servers.GetJob)
		group.PATCH("/:organization/servers/:server", servers.Update)
		group.DELETE("/:organization/servers/:server", servers.Delete)
	}

	group = router.Group("/agent", middleware.RequireAgent())
	{
		group.POST("/heartbeat", agent.Heartbeat)
		group.GET("/jobs/next", agent.NextJob)
		group.POST("/jobs/:job/result", agent.ReportJob)
	}
}
