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
	var req requests.LoginRequest

	if !bindJSON(ctx, &req) {
		return
	}

	user, err := (models.User{}).FindByEmail(ctx, req.Email)
	if err != nil || !helpers.CheckPassword(req.Password, user.Password) {
		respond(ctx).fail(
			"The provided credentials doesn't match our system",
			gin.H{},
			http.StatusUnauthorized,
		)
		return
	}

	token, err := helpers.GenerateToken()
	if err != nil {
		respond(ctx).fail("Failed to login", gin.H{}, http.StatusInternalServerError)
		return
	}

	if err := (&models.PersonalAccessToken{
		UserID: int(user.ID),
		Token:  token,
	}).Create(ctx); err != nil {
		respond(ctx).fail("Failed to login", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Login successful!", gin.H{
		"access_token": token,
	}, 200)
}

func (AuthController) Register(ctx *gin.Context) {
	var req requests.RegisterRequest

	if !bindJSON(ctx, &req) {
		return
	}

	if (models.User{}).IsEmailTaken(ctx, req.Email) {
		respond(ctx).fail(
			"User already exists with the given email",
			gin.H{
				"email": "The email has already been taken",
			},
			http.StatusConflict,
		)
		return
	}

	hashedPassword, err := helpers.HashPassword(req.Password)
	if err != nil {
		respond(ctx).fail("Failed to register user", gin.H{}, http.StatusInternalServerError)
		return
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := user.Create(ctx); err != nil {
		respond(ctx).fail("Failed to register user", gin.H{}, http.StatusInternalServerError)
		return
	}

	slug, err := (models.Organization{}).GenerateUniqueSlug(ctx, user.Name)
	if err != nil {
		respond(ctx).fail("Failed to register user", gin.H{}, http.StatusInternalServerError)
		return
	}

	if err := (&models.Organization{
		Name:   user.Name,
		Slug:   slug,
		UserID: user.ID,
	}).Create(ctx); err != nil {
		respond(ctx).fail("Failed to register user", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("User registration successful", gin.H{}, 201)
}

func (AuthController) Me(ctx *gin.Context) {
	respond(ctx).ok(
		"User fetched successfully!",
		resources.User(currentUser(ctx)),
		http.StatusOK,
	)
}
