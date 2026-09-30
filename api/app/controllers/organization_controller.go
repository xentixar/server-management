package controllers

import (
	"api/app/models"
	"api/app/resources"
	"api/config"

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
