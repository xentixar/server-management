package controllers

import (
	"api/app/models"
	"api/app/policies"
	"api/app/requests"
	"api/app/resources"

	"github.com/gin-gonic/gin"
)

type OrganizationController struct{}

func (OrganizationController) List(ctx *gin.Context) {
	user := currentUser(ctx)

	organizations, err := (models.Organization{}).FindByMember(ctx, user.ID)
	if err != nil {
		respond(ctx).fail("Failed to fetch organizations", gin.H{}, 500)
		return
	}

	roles, err := (models.OrganizationMember{}).RolesByUser(ctx, user.ID)
	if err != nil {
		respond(ctx).fail("Failed to fetch organization roles", gin.H{}, 500)
		return
	}

	respond(ctx).ok(
		"Organization fetched successfully!",
		resources.OrganizationCollection(organizations, roles),
		200,
	)
}

func (OrganizationController) Create(ctx *gin.Context) {
	user := currentUser(ctx)
	if !denyUnless(ctx, (policies.OrganizationPolicy{}).Create(user)) {
		return
	}

	var req requests.OrganizationRequest
	if !bindJSON(ctx, &req) {
		return
	}

	slug, err := (models.Organization{}).GenerateUniqueSlug(ctx, req.Name)
	if err != nil {
		respond(ctx).fail("Failed to generate organization slug", gin.H{}, 500)
		return
	}

	organization := models.Organization{
		Name:   req.Name,
		Slug:   slug,
		UserID: user.ID,
	}

	if err := organization.Create(ctx); err != nil {
		respond(ctx).fail("Failed to create organization", gin.H{}, 500)
		return
	}

	respond(ctx).ok(
		"Organization created successfully!",
		resources.OrganizationResource(organization, "owner"),
		201,
	)
}

func (OrganizationController) Get(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}

	organization, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.OrganizationPolicy{}).View(member)) {
		return
	}

	respond(ctx).ok(
		"Organization fetched successfully!",
		resources.OrganizationResource(organization, member.Role),
		200,
	)
}

func (OrganizationController) Update(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}

	organization, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.OrganizationPolicy{}).Update(member)) {
		return
	}

	var req requests.OrganizationRequest
	if !bindJSON(ctx, &req) {
		return
	}

	if organization.Name != req.Name {
		slug, err := (models.Organization{}).GenerateUniqueSlug(ctx, req.Name)
		if err != nil {
			respond(ctx).fail("Failed to generate organization slug", gin.H{}, 500)
			return
		}
		organization.Slug = slug
	}

	organization.Name = req.Name

	if err := organization.Update(ctx); err != nil {
		respond(ctx).fail("Failed to update organization", gin.H{}, 500)
		return
	}

	respond(ctx).ok(
		"Organization updated successfully!",
		resources.OrganizationResource(organization, member.Role),
		200,
	)
}

func (OrganizationController) Delete(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}

	organization, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.OrganizationPolicy{}).Delete(member)) {
		return
	}

	if err := organization.Delete(ctx); err != nil {
		respond(ctx).fail("Failed to delete organization", gin.H{}, 500)
		return
	}

	respond(ctx).ok("Organization deleted successfully!", gin.H{}, 200)
}
