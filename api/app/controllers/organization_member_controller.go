package controllers

import (
	"strconv"

	"api/app/models"
	"api/app/resources"

	"github.com/gin-gonic/gin"
)

type OrganizationMemberController struct{}

func (OrganizationMemberController) Index(ctx *gin.Context) {
	organizationID, err := strconv.Atoi(ctx.Param("organization"))
	if err != nil {
		response(ctx).error("Invalid organization ID", gin.H{}, 400)
		return
	}

	user := currentUser(ctx)

	organization, err := (models.Organization{}).FindByID(ctx, uint(organizationID), user.ID, "Members", "Members.User")
	if err != nil {
		response(ctx).error("Organization not found", gin.H{}, 404)
		return
	}

	organizationMembers := make([]map[string]any, 0, len(organization.Members))
	for _, organizationMember := range organization.Members {
		organizationMembers = append(
			organizationMembers,
			resources.OrganizationMemberResource(organizationMember),
		)
	}

	response(ctx).success("Organization members fetched successfully!", organizationMembers, 200)
}
