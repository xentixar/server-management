package controllers

import (
	"strconv"

	"api/app/models"
	"api/app/resources"

	"github.com/gin-gonic/gin"
)

type OrganizationMemberController struct{}

func (OrganizationMemberController) List(ctx *gin.Context) {
	organizationID, err := strconv.Atoi(ctx.Param("organization"))
	if err != nil {
		respond(ctx).fail("Invalid organization ID", gin.H{}, 400)
		return
	}

	user := currentUser(ctx)

	organization, err := (models.Organization{}).FindOwnedByID(ctx, uint(organizationID), user.ID, "Members", "Members.User")
	if err != nil {
		respond(ctx).fail("Organization not found", gin.H{}, 404)
		return
	}

	payload := make([]map[string]any, 0, len(organization.Members))
	for _, member := range organization.Members {
		payload = append(payload, resources.Member(member))
	}

	respond(ctx).ok("Organization members fetched successfully!", payload, 200)
}
