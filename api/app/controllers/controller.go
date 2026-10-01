// Package controllers handles HTTP request and response logic for API endpoints.
package controllers

import (
	"net/http"

	"api/app/models"
	"api/app/validators"

	"github.com/gin-gonic/gin"
)

type Response struct {
	context *gin.Context
}

func response(context *gin.Context) Response {
	return Response{
		context: context,
	}
}

func (r Response) success(message string, data any, code int) {
	r.context.JSON(code, map[string]any{
		"message": message,
		"data":    data,
		"success": true,
	})
}

func (r Response) error(message string, errors any, code int) {
	r.context.JSON(code, map[string]any{
		"message": message,
		"data":    errors,
		"success": false,
	})
}

func validate(context *gin.Context, request any) bool {
	if err := context.ShouldBindJSON(request); err != nil {
		fieldErrors := validators.FormatValidationErrors(err, request)

		response(context).error(
			"Validation error",
			fieldErrors,
			http.StatusBadRequest,
		)

		return false
	}

	return true
}

func currentUser(ctx *gin.Context) models.User {
	return ctx.MustGet("user").(models.User)
}
