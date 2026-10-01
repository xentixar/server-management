// Package controllers handles HTTP request and response logic for API endpoints.
package controllers

import (
	"net/http"
	"strconv"

	"api/app/models"
	"api/app/validators"

	"github.com/gin-gonic/gin"
)

type apiResponse struct {
	ctx *gin.Context
}

func respond(ctx *gin.Context) apiResponse {
	return apiResponse{ctx: ctx}
}

func (r apiResponse) ok(message string, data any, code int) {
	r.ctx.JSON(code, map[string]any{
		"message": message,
		"data":    data,
		"success": true,
	})
}

func (r apiResponse) fail(message string, data any, code int) {
	r.ctx.JSON(code, map[string]any{
		"message": message,
		"data":    data,
		"success": false,
	})
}

func bindJSON(ctx *gin.Context, request any) bool {
	if err := ctx.ShouldBindJSON(request); err != nil {
		respond(ctx).fail(
			"Validation error",
			validators.FormatValidationErrors(err, request),
			http.StatusBadRequest,
		)
		return false
	}

	return true
}

func currentUser(ctx *gin.Context) models.User {
	return ctx.MustGet("user").(models.User)
}

func uintParam(ctx *gin.Context, name, invalidMessage string) (uint, bool) {
	id, err := strconv.Atoi(ctx.Param(name))
	if err != nil || id <= 0 {
		respond(ctx).fail(invalidMessage, gin.H{}, http.StatusBadRequest)
		return 0, false
	}

	return uint(id), true
}

func requireOrgMember(ctx *gin.Context, organizationID uint, withRelations ...string) (models.Organization, models.OrganizationMember, bool) {
	user := currentUser(ctx)

	organization, err := (models.Organization{}).FindForMember(ctx, organizationID, user.ID, withRelations...)
	if err != nil {
		respond(ctx).fail("Organization not found", gin.H{}, http.StatusNotFound)
		return models.Organization{}, models.OrganizationMember{}, false
	}

	member, err := (models.OrganizationMember{}).FindByUser(ctx, organizationID, user.ID)
	if err != nil || member == nil {
		respond(ctx).fail("Organization not found", gin.H{}, http.StatusNotFound)
		return models.Organization{}, models.OrganizationMember{}, false
	}

	return organization, *member, true
}

func denyUnless(ctx *gin.Context, allowed bool) bool {
	if allowed {
		return true
	}
	respond(ctx).fail("Forbidden", gin.H{}, http.StatusForbidden)
	return false
}
