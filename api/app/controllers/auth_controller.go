package controllers

import (
	"context"
	"net/http"

	"api/app/models"
	"api/app/requests"
	"api/config"
	"api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct{}

func (AuthController) Login(ctx *gin.Context) {
	var loginRequest requests.LoginRequest

	if !validate(ctx, &loginRequest) {
		return
	}

	bgContext := context.Background()

	user, err := gorm.G[models.User](config.GetDatabase()).Where("email = ?", loginRequest.Email).First(bgContext)
	if err != nil {
		response(ctx).error(
			"The provided credentials doesn't match our system",
			gin.H{},
			http.StatusUnauthorized,
		)
		return
	}

	if !config.CheckPasswordHash(loginRequest.Password, user.Password) {
		response(ctx).error(
			"The provided credentials doesn't match our system",
			gin.H{},
			http.StatusUnauthorized,
		)
		return
	}

	token, _ := utils.GenerateToken()

	_ = gorm.G[models.PersonalAccessToken](config.GetDatabase()).Create(bgContext, &models.PersonalAccessToken{
		UserID: int(user.ID),
		Token:  token,
	})

	response(ctx).success("Login successfull!", gin.H{
		"access_token": token,
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

	hashedPassword, _ := config.HashPassword(registerRequest.Password)

	_ = gorm.G[models.User](config.GetDatabase()).Create(bgContext, &models.User{
		Name:     registerRequest.Name,
		Email:    registerRequest.Email,
		Password: hashedPassword,
	})

	response(ctx).success("User registration successful", gin.H{}, 201)
}

func (AuthController) GetUser(ctx *gin.Context) {
	user, _ := ctx.Get("user")
	response(ctx).success(
		"User fetched successfully!",
		user,
		http.StatusOK,
	)
}
