// Package controllers contains all the controllers
package controllers

import "github.com/gin-gonic/gin"

type Response struct {
	success func(message string, data any, code int)
	error   func(message string, errors any, code int)
}

var context *gin.Context

func response(ctx *gin.Context) Response {
	context = ctx
	res := Response{
		success: success,
		error:   error,
	}

	return res
}

func success(message string, data any, code int) {
	context.JSON(code, map[string]any{
		"message": message,
		"data":    data,
		"success": true,
	})
}

func error(message string, errors any, code int) {
	context.JSON(code, map[string]any{
		"message": message,
		"data":    errors,
		"success": false,
	})
}
