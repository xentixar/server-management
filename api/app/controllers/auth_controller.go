package controllers

import (
	"net/http"

	"api/app/helpers"
	"api/app/models"
	"api/app/requests"
	"api/app/resources"

	"github.com/gin-gonic/gin"
)

type AuthController struct{}

func (AuthController) Login(ctx *gin.Context) {
	var loginRequest requests.LoginRequest

	if !validate(ctx, &loginRequest) {
		return
	}

	user, err := (models.User{}).FindByEmail(ctx, loginRequest.Email)
	if err != nil || !helpers.CheckPasswordHash(loginRequest.Password, user.Password) {
		response(ctx).error(
			"The provided credentials doesn't match our system",
			gin.H{},
			http.StatusUnauthorized,
		)
		return
	}

	token, err := helpers.GenerateToken()
	if err != nil {
		response(ctx).error("Failed to login", gin.H{}, http.StatusInternalServerError)
		return
	}

	if err := (&models.PersonalAccessToken{
		UserID: int(user.ID),
		Token:  token,
	}).Create(ctx); err != nil {
		response(ctx).error("Failed to login", gin.H{}, http.StatusInternalServerError)
		return
	}

	response(ctx).success("Login successful!", gin.H{
		"access_token": token,
	}, 200)
}

func (AuthController) Register(ctx *gin.Context) {
	var registerRequest requests.RegisterRequest

	if !validate(ctx, &registerRequest) {
		return
	}

	if (models.User{}).IsEmailTaken(ctx, registerRequest.Email) {
		response(ctx).error(
			"User already exists with the given email",
			gin.H{
				"email": "The email has already been taken",
			},
			http.StatusConflict,
		)
		return
	}

	hashedPassword, err := helpers.HashPassword(registerRequest.Password)
	if err != nil {
		response(ctx).error("Failed to register user", gin.H{}, http.StatusInternalServerError)
		return
	}

	user := models.User{
		Name:     registerRequest.Name,
		Email:    registerRequest.Email,
		Password: hashedPassword,
	}

	if err := user.Create(ctx); err != nil {
		response(ctx).error("Failed to register user", gin.H{}, http.StatusInternalServerError)
		return
	}

	slug, err := (models.Organization{}).UniqueSlug(ctx, user.Name)
	if err != nil {
		response(ctx).error("Failed to register user", gin.H{}, http.StatusInternalServerError)
		return
	}

	if err := (&models.Organization{
		Name:   user.Name,
		Slug:   slug,
		UserID: user.ID,
	}).Create(ctx); err != nil {
		response(ctx).error("Failed to register user", gin.H{}, http.StatusInternalServerError)
		return
	}

	response(ctx).success("User registration successful", gin.H{}, 201)
}

func (AuthController) GetUser(ctx *gin.Context) {
	response(ctx).success(
		"User fetched successfully!",
		resources.UserResource(currentUser(ctx)),
		http.StatusOK,
	)
}
