package controllers

import (
	"net/http"

	"api/app/requests"
	"api/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct{}

func (AuthController) Login(context *gin.Context) {
	var loginRequest requests.LoginRequest

	if err := context.ShouldBindJSON(&loginRequest); err != nil {
		fieldErrors := utils.FormatValidationErrors(err, loginRequest)
		response(context).error("Validation error", fieldErrors, http.StatusBadRequest)

		return
	}

	email := loginRequest.Email
	password := loginRequest.Password

	response(context).success("Login successful!", gin.H{
		"email":    email,
		"password": password,
	}, 200)
}
