package controllers

import (
	"strconv"

	"api/app/models"
	"api/app/requests"
	"api/app/resources"

	"github.com/gin-gonic/gin"
)

type OrganizationController struct{}

func (OrganizationController) List(ctx *gin.Context) {
	user := currentUser(ctx)

	organizations, err := (models.Organization{}).FindOwnedByUser(ctx, user.ID)
	if err != nil {
		respond(ctx).fail("Failed to fetch organizations", gin.H{}, 500)
		return
	}

	payload := make([]map[string]any, 0, len(organizations))
	for _, organization := range organizations {
		payload = append(payload, resources.Organization(organization))
	}

	respond(ctx).ok("Organization fetched successfully!", payload, 200)
}

func (OrganizationController) Create(ctx *gin.Context) {
	var req requests.OrganizationRequest

	if !bindJSON(ctx, &req) {
		return
	}

	user := currentUser(ctx)

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
		resources.Organization(organization),
		201,
	)
}

func (OrganizationController) Get(ctx *gin.Context) {
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

	respond(ctx).ok("Organization fetched successfully!", resources.Organization(organization), 200)
}

func (OrganizationController) Update(ctx *gin.Context) {
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
		resources.Organization(organization),
		200,
	)
}

func (OrganizationController) Delete(ctx *gin.Context) {
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

	if err := organization.Delete(ctx); err != nil {
		respond(ctx).fail("Failed to delete organization", gin.H{}, 500)
		return
	}

	respond(ctx).ok("Organization deleted successfully!", gin.H{}, 200)
}
