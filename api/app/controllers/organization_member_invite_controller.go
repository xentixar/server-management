package controllers

import (
	"strconv"

	"api/app/constants"
	"api/app/models"
	"api/app/requests"

	"github.com/gin-gonic/gin"
)

type OrganizationMemberInviteController struct{}

func (OrganizationMemberInviteController) Invite(ctx *gin.Context) {
	organizationID, err := strconv.Atoi(ctx.Param("organization"))
	if err != nil {
		response(ctx).error("Invalid organization ID", gin.H{}, 400)
		return
	}

	user := currentUser(ctx)

	organization, err := (models.Organization{}).FindByID(ctx, uint(organizationID), user.ID)
	if err != nil {
		response(ctx).error("Organization not found", gin.H{}, 404)
		return
	}

	var req requests.OrganizationMemberInviteRequest
	if !validate(ctx, &req) {
		return
	}

	role, ok := constants.ROLES[req.Role]
	if !ok {
		response(ctx).error("Invalid role", gin.H{}, 422)
		return
	}

	member, err := (models.OrganizationMember{}).FindOneByEmail(ctx, organization.ID, req.Email)
	if err != nil {
		response(ctx).error("Failed to check existing member", gin.H{}, 500)
		return
	}
	if member != nil {
		response(ctx).error("This email is already a member", gin.H{}, 422)
		return
	}

	invite, err := (models.OrganizationMemberInvite{}).FindOneByEmail(ctx, organization.ID, req.Email)
	if err != nil {
		response(ctx).error("Failed to check existing invite", gin.H{}, 500)
		return
	}
	if invite != nil {
		response(ctx).error("This email is already invited", gin.H{}, 422)
		return
	}

	if err := (models.OrganizationMemberInvite{}).Create(ctx, organization.ID, req.Email, role); err != nil {
		response(ctx).error("Error occurred while inviting user", gin.H{}, 500)
		return
	}

	response(ctx).success("User invited successfully!", gin.H{}, 201)
}
