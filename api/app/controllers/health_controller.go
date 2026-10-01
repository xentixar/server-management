package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func (HealthController) Check(ctx *gin.Context) {
	respond(ctx).ok("Healthy", gin.H{}, http.StatusOK)
}
