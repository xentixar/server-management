package controllers

import (
	"api/app/models"
	"api/app/requests"
	"api/app/resources"
	"api/config"
	"api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrganizationController struct{}

func (OrganizationController) Index(ctx *gin.Context) {
	value, _ := ctx.Get("user")
	user := value.(models.User)

	organizations, err := gorm.
		G[models.Organization](config.GetDatabase()).
		Where("user_id = ?", user.ID).
		Find(ctx)
	if err != nil {
		ctx.JSON(500, gin.H{
			"message": "Failed to fetch organizations",
		})
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

	value, exists := ctx.Get("user")
	if !exists {
		response(ctx).error("Unauthenticated", gin.H{}, 401)
		return
	}

	user, ok := value.(models.User)
	if !ok {
		response(ctx).error("Invalid user", gin.H{}, 401)
		return
	}

	baseSlug := utils.Slugify(storeOrganizationRequest.Name)

	var slug string

	for {
		randomSuffix, err := utils.RandomAlphanumeric(5)
		if err != nil {
			response(ctx).error("Failed to generate organization slug", gin.H{}, 500)
			return
		}

		slug = baseSlug + "-" + randomSuffix

		if (models.Organization{}).IsSlugAvailable(slug) {
			break
		}
	}

	organization := models.Organization{
		Name:   storeOrganizationRequest.Name,
		Slug:   slug,
		UserID: user.ID,
	}

	result := gorm.WithResult()

	if err := gorm.
		G[models.Organization](config.GetDatabase(), result).
		Create(ctx, &organization); err != nil {
		response(ctx).error("Failed to create organization", gin.H{}, 500)
		return
	}

	response(ctx).success(
		"Organization created successfully!",
		resources.OrganizationResource(organization),
		201,
	)
}
