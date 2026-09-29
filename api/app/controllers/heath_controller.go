package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func (HealthController) GetHealth(context *gin.Context) {
	response(context).success("Healthy", gin.H{}, http.StatusOK)
}
