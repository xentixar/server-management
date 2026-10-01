package controllers

import (
	"strconv"

	"api/app/models"
	"api/app/requests"
	"api/app/resources"

	"github.com/gin-gonic/gin"
)

type OrganizationController struct{}

func (OrganizationController) Index(ctx *gin.Context) {
	user := currentUser(ctx)

	organizations, err := (models.Organization{}).FindByUserID(ctx, user.ID)
	if err != nil {
		response(ctx).error("Failed to fetch organizations", gin.H{}, 500)
		return
	}

	organizationsResponse := make([]map[string]any, 0, len(organizations))
	for _, organization := range organizations {
		organizationsResponse = append(
			organizationsResponse,
			resources.OrganizationResource(organization),
		)
	}

	response(ctx).success("Organization fetched successfully!", organizationsResponse, 200)
}

func (OrganizationController) Store(ctx *gin.Context) {
	var storeOrganizationRequest requests.StoreOrganizationRequest

	if !validate(ctx, &storeOrganizationRequest) {
		return
	}

	user := currentUser(ctx)

	slug, err := (models.Organization{}).UniqueSlug(ctx, storeOrganizationRequest.Name)
	if err != nil {
		response(ctx).error("Failed to generate organization slug", gin.H{}, 500)
		return
	}

	organization := models.Organization{
		Name:   storeOrganizationRequest.Name,
		Slug:   slug,
		UserID: user.ID,
	}

	if err := organization.Create(ctx); err != nil {
		response(ctx).error("Failed to create organization", gin.H{}, 500)
		return
	}

	response(ctx).success(
		"Organization created successfully!",
		resources.OrganizationResource(organization),
		201,
	)
}

func (OrganizationController) Show(ctx *gin.Context) {
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

	response(ctx).success("Organization fetched successfully!", resources.OrganizationResource(organization), 200)
}

func (OrganizationController) Update(ctx *gin.Context) {
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

	var storeOrganizationRequest requests.StoreOrganizationRequest

	if !validate(ctx, &storeOrganizationRequest) {
		return
	}

	if organization.Name != storeOrganizationRequest.Name {
		slug, err := (models.Organization{}).UniqueSlug(ctx, storeOrganizationRequest.Name)
		if err != nil {
			response(ctx).error("Failed to generate organization slug", gin.H{}, 500)
			return
		}
		organization.Slug = slug
	}

	organization.Name = storeOrganizationRequest.Name

	if err := organization.Update(ctx); err != nil {
		response(ctx).error("Failed to update organization", gin.H{}, 500)
		return
	}

	response(ctx).success(
		"Organization updated successfully!",
		resources.OrganizationResource(organization),
		200,
	)
}

func (OrganizationController) Destroy(ctx *gin.Context) {
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

	if err := organization.Delete(ctx); err != nil {
		response(ctx).error("Failed to delete organization", gin.H{}, 500)
		return
	}

	response(ctx).success("Organization deleted successfully!", gin.H{}, 200)
}
