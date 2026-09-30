package controllers

import (
	"strconv"

	"api/app/models"
	"api/app/requests"
	"api/app/resources"
	"api/config"

	"github.com/gin-gonic/gin"
)

type OrganizationController struct{}

func (OrganizationController) Index(ctx *gin.Context) {
	user := currentUser(ctx)

	organizations, err := (models.Organization{}).FindByUserID(user.ID)
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

	slug, err := (models.Organization{}).UniqueSlug(storeOrganizationRequest.Name)
	if err != nil {
		response(ctx).error("Failed to generate organization slug", gin.H{}, 500)
		return
	}

	organization := models.Organization{
		Name:   storeOrganizationRequest.Name,
		Slug:   slug,
		UserID: user.ID,
	}

	if err := organization.Create(); err != nil {
		response(ctx).error("Failed to create organization", gin.H{}, 500)
		return
	}

	response(ctx).success(
		"Organization created successfully!",
		resources.OrganizationResource(organization),
		201,
	)
}

func (OrganizationController) Update(ctx *gin.Context) {
	organizationID, err := strconv.Atoi(ctx.Param("organization"))
	if err != nil {
		response(ctx).error("Invalid organization ID", gin.H{}, 400)
		return
	}

	user := currentUser(ctx)

	organization, err := (models.Organization{}).FindByID(uint(organizationID), user.ID)
	if err != nil {
		response(ctx).error("Organization not found", gin.H{}, 404)
		return
	}

	var storeOrganizationRequest requests.StoreOrganizationRequest

	if !validate(ctx, &storeOrganizationRequest) {
		return
	}

	slug, err := (models.Organization{}).UniqueSlug(storeOrganizationRequest.Name)
	if err != nil {
		response(ctx).error("Failed to generate organization slug", gin.H{}, 500)
		return
	}

	organization.Name = storeOrganizationRequest.Name
	organization.Slug = slug

	config.GetDatabase().Save(organization)

	response(ctx).success(
		"Organization updated successfully!",
		resources.OrganizationResource(organization),
		200,
	)
}
