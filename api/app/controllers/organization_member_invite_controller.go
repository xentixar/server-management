package controllers

import (
	"api/app/constants"
	"api/app/models"
	"api/app/policies"
	"api/app/requests"
	"api/app/resources"

	"github.com/gin-gonic/gin"
)

type OrganizationMemberInviteController struct{}

func (OrganizationMemberInviteController) List(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}

	_, actor, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.MemberPolicy{}).ViewInvites(actor)) {
		return
	}

	invites, err := (models.OrganizationMemberInvite{}).FindByOrganization(ctx, organizationID)
	if err != nil {
		respond(ctx).fail("Failed to fetch invites", gin.H{}, 500)
		return
	}

	respond(ctx).ok(
		"Organization invites fetched successfully!",
		resources.OrganizationMemberInviteCollection(invites),
		200,
	)
}

func (OrganizationMemberInviteController) Create(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}

	organization, actor, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.MemberPolicy{}).Invite(actor)) {
		return
	}

	var req requests.InviteMemberRequest
	if !bindJSON(ctx, &req) {
		return
	}

	if req.Role == constants.OwnerRole {
		respond(ctx).fail("Cannot invite user as owner", gin.H{}, 422)
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

func (OrganizationMemberInviteController) Accept(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}

	user := currentUser(ctx)

	if _, err := (models.Organization{}).FindByID(ctx, organizationID); err != nil {
		respond(ctx).fail("Organization not found", gin.H{}, 404)
		return
	}

	invite, err := (models.OrganizationMemberInvite{}).FindByEmail(ctx, organizationID, user.Email)
	if err != nil {
		respond(ctx).fail("Failed to fetch invite", gin.H{}, 500)
		return
	}
	if invite == nil {
		respond(ctx).fail("Invite not found", gin.H{}, 404)
		return
	}

	role, ok := constants.Roles[invite.Role]
	if !ok {
		respond(ctx).fail("Invite has an invalid role", gin.H{}, 422)
		return
	}

	if err := (models.OrganizationMember{}).Upsert(ctx, organizationID, user.ID, role); err != nil {
		respond(ctx).fail("Failed to accept invite", gin.H{}, 500)
		return
	}

	if err := invite.Delete(ctx); err != nil {
		respond(ctx).fail("Failed to clear invite", gin.H{}, 500)
		return
	}

	respond(ctx).ok("Invite accepted successfully!", gin.H{}, 200)
}

func (OrganizationMemberInviteController) Delete(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}

	inviteID, ok := uintParam(ctx, "invite", "Invalid invite ID")
	if !ok {
		return
	}

	_, actor, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.MemberPolicy{}).RevokeInvite(actor)) {
		return
	}

	invite, err := (models.OrganizationMemberInvite{}).FindByID(ctx, inviteID, organizationID)
	if err != nil {
		respond(ctx).fail("Invite not found", gin.H{}, 404)
		return
	}

	if err := invite.Delete(ctx); err != nil {
		respond(ctx).fail("Failed to revoke invite", gin.H{}, 500)
		return
	}

	respond(ctx).ok("Invite revoked successfully!", gin.H{}, 200)
}
