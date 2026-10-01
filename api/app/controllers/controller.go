// Package controllers handles HTTP request and response logic for API endpoints.
package controllers

import (
	"net/http"

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
