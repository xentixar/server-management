package controllers

import (
	"context"
	"net/http"

	"api/app/models"
	"api/app/requests"
	"api/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct{}

func (AuthController) Login(ctx *gin.Context) {
	var loginRequest requests.LoginRequest

	if !validate(ctx, &loginRequest) {
		return
	}

	email := loginRequest.Email
	password := loginRequest.Password

	response(ctx).success("Login successful!", gin.H{
		"email":    email,
		"password": password,
	}, 200)
}

func (AuthController) Register(ctx *gin.Context) {
	var registerRequest requests.RegisterRequest

	if !validate(ctx, &registerRequest) {
		return
	}

	bgContext := context.Background()

	_, err := gorm.G[models.User](config.GetDatabase()).Where("email = ?", registerRequest.Email).First(bgContext)

	if err == nil {
		response(ctx).error(
			"User already exists with the given email",
			gin.H{
				"email": "The email has already been taken",
			},
			http.StatusConflict,
		)
		return
	}

	_ = gorm.G[models.User](config.GetDatabase()).Create(bgContext, &models.User{
		Name:     registerRequest.Name,
		Email:    registerRequest.Email,
		Password: registerRequest.Password,
	})

	response(ctx).success("User registration successful", gin.H{}, 201)
}
