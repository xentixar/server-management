package controllers

import (
	"api/app/constants"
	"api/app/models"
	"api/app/requests"
	"api/app/resources"

	"github.com/gin-gonic/gin"
)

type OrganizationMemberController struct{}

func (OrganizationMemberController) List(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}

	user := currentUser(ctx)

	organization, err := (models.Organization{}).FindOwnedByID(ctx, organizationID, user.ID, "Members", "Members.User")
	if err != nil {
		respond(ctx).fail("Organization not found", gin.H{}, 404)
		return
	}

	respond(ctx).ok(
		"Organization members fetched successfully!",
		resources.OrganizationMemberCollection(organization.Members),
		200,
	)
}

func (OrganizationMemberController) Update(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}

	memberID, ok := uintParam(ctx, "member", "Invalid member ID")
	if !ok {
		return
	}

	user := currentUser(ctx)

	if _, err := (models.Organization{}).FindOwnedByID(ctx, organizationID, user.ID); err != nil {
		respond(ctx).fail("Organization not found", gin.H{}, 404)
		return
	}

	var req requests.UpdateMemberRoleRequest
	if !bindJSON(ctx, &req) {
		return
	}

	if req.Role == constants.OwnerRole {
		respond(ctx).fail("Cannot assign owner role", gin.H{}, 422)
		return
	}

	role, ok := constants.Roles[req.Role]
	if !ok {
		respond(ctx).fail("Invalid role", gin.H{}, 422)
		return
	}

	member, err := (models.OrganizationMember{}).FindByID(ctx, memberID, organizationID, "User")
	if err != nil {
		respond(ctx).fail("Member not found", gin.H{}, 404)
		return
	}

	if err := member.UpdateRole(ctx, role); err != nil {
		respond(ctx).fail(err.Error(), gin.H{}, 422)
		return
	}

	respond(ctx).ok(
		"Member role updated successfully!",
		resources.OrganizationMemberResource(member),
		200,
	)
}

func (OrganizationMemberController) Delete(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}

	memberID, ok := uintParam(ctx, "member", "Invalid member ID")
	if !ok {
		return
	}

	user := currentUser(ctx)

	if _, err := (models.Organization{}).FindOwnedByID(ctx, organizationID, user.ID); err != nil {
		respond(ctx).fail("Organization not found", gin.H{}, 404)
		return
	}

	member, err := (models.OrganizationMember{}).FindByID(ctx, memberID, organizationID)
	if err != nil {
		respond(ctx).fail("Member not found", gin.H{}, 404)
		return
	}

	if err := member.Delete(ctx); err != nil {
		respond(ctx).fail(err.Error(), gin.H{}, 422)
		return
	}

	respond(ctx).ok("Member removed successfully!", gin.H{}, 200)
}
