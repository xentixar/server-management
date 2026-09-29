package controllers

import (
	"api/app/requests"

	"github.com/gin-gonic/gin"
)

type AuthController struct{}

func (AuthController) Login(context *gin.Context) {
	var loginRequest requests.LoginRequest

	if !validate(context, &loginRequest) {
		return
	}

	email := loginRequest.Email
	password := loginRequest.Password

	response(context).success("Login successful!", gin.H{
		"email":    email,
		"password": password,
	}, 200)
}

func (AuthController) Register(context *gin.Context) {
	var registerRequest requests.RegisterRequest

	if !validate(context, &registerRequest) {
		return
	}
}
