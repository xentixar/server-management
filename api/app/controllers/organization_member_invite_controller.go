package controllers

import (
	"strconv"

	"api/app/constants"
	"api/app/models"
	"api/app/requests"

	"github.com/gin-gonic/gin"
)

type OrganizationMemberInviteController struct{}

func (OrganizationMemberInviteController) Create(ctx *gin.Context) {
	organizationID, err := strconv.Atoi(ctx.Param("organization"))
	if err != nil {
		respond(ctx).fail("Invalid organization ID", gin.H{}, 400)
		return
	}

	user := currentUser(ctx)

	organization, err := (models.Organization{}).FindOwnedByID(ctx, uint(organizationID), user.ID)
	if err != nil {
		respond(ctx).fail("Organization not found", gin.H{}, 404)
		return
	}

	var req requests.InviteMemberRequest
	if !bindJSON(ctx, &req) {
		return
	}

	role, ok := constants.Roles[req.Role]
	if !ok {
		respond(ctx).fail("Invalid role", gin.H{}, 422)
		return
	}

	member, err := (models.OrganizationMember{}).FindByEmail(ctx, organization.ID, req.Email)
	if err != nil {
		respond(ctx).fail("Failed to check existing member", gin.H{}, 500)
		return
	}
	if member != nil {
		respond(ctx).fail("This email is already a member", gin.H{}, 422)
		return
	}

	invite, err := (models.OrganizationMemberInvite{}).FindByEmail(ctx, organization.ID, req.Email)
	if err != nil {
		respond(ctx).fail("Failed to check existing invite", gin.H{}, 500)
		return
	}
	if invite != nil {
		respond(ctx).fail("This email is already invited", gin.H{}, 422)
		return
	}

	if err := (models.OrganizationMemberInvite{}).Create(ctx, organization.ID, req.Email, role); err != nil {
		respond(ctx).fail("Error occurred while inviting user", gin.H{}, 500)
		return
	}

	respond(ctx).ok("User invited successfully!", gin.H{}, 201)
}
