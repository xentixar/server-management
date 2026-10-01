package controllers

import (
	"net/http"

	"api/app/models"
	"api/app/requests"
	"api/app/resources"
	"api/config"
	"api/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct{}

func (AuthController) Login(ctx *gin.Context) {
	var loginRequest requests.LoginRequest

	if !validate(ctx, &loginRequest) {
		return
	}

	user, err := (models.User{}).FindByEmail(loginRequest.Email)
	if err != nil || !config.CheckPasswordHash(loginRequest.Password, user.Password) {
		response(ctx).error(
			"The provided credentials doesn't match our system",
			gin.H{},
			http.StatusUnauthorized,
		)
		return
	}

	token, _ := utils.GenerateToken()

	_ = (&models.PersonalAccessToken{
		UserID: int(user.ID),
		Token:  token,
	}).Create()

	response(ctx).success("Login successful!", gin.H{
		"access_token": token,
	}, 200)
}

func (AuthController) Register(ctx *gin.Context) {
	var registerRequest requests.RegisterRequest

	if !validate(ctx, &registerRequest) {
		return
	}

	if (models.User{}).IsEmailTaken(registerRequest.Email) {
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

	user := models.User{
		Name:     registerRequest.Name,
		Email:    registerRequest.Email,
		Password: hashedPassword,
	}

	_ = user.Create()

	slug, _ := (models.Organization{}).UniqueSlug(user.Name)

	_ = (&models.Organization{
		Name:   user.Name,
		Slug:   slug,
		UserID: user.ID,
	}).Create()

	response(ctx).success("User registration successful", gin.H{}, 201)
}

func (AuthController) GetUser(ctx *gin.Context) {
	response(ctx).success(
		"User fetched successfully!",
		resources.UserResource(currentUser(ctx)),
		http.StatusOK,
	)
}
