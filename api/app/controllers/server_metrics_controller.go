package controllers

import (
	"errors"
	"net/http"
	"time"

	"api/app/models"
	"api/app/resources"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (ServerController) GetMetrics(ctx *gin.Context) {
	server, ok := loadServerForView(ctx)
	if !ok {
		return
	}

	rangeKey := ctx.DefaultQuery("range", "1h")
	from, bucket, okRange := metricsRange(rangeKey)
	if !okRange {
		respond(ctx).fail("Invalid range", gin.H{"range": "use 1h, 6h, 24h, or 7d"}, http.StatusBadRequest)
		return
	}
	to := time.Now().UTC()

	latest, err := (models.ServerMetricSample{}).Latest(ctx, server.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		respond(ctx).fail("Failed to load metrics", gin.H{}, http.StatusInternalServerError)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		latest = nil
	}

	series, err := (models.ServerMetricSample{}).Series(ctx, server.ID, from, to, bucket)
	if err != nil {
		respond(ctx).fail("Failed to load metrics", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Metrics fetched", gin.H{
		"range":  rangeKey,
		"from":   from.Format("2006-01-02T15:04:05Z"),
		"to":     to.Format("2006-01-02T15:04:05Z"),
		"latest": resources.ServerMetricSampleResource(latest),
		"series": resources.ServerMetricSeriesResource(series),
	}, http.StatusOK)
}

func metricsRange(key string) (from time.Time, bucket string, ok bool) {
	now := time.Now().UTC()
	switch key {
	case "1h":
		return now.Add(-1 * time.Hour), "", true
	case "6h":
		return now.Add(-6 * time.Hour), "", true
	case "24h":
		return now.Add(-24 * time.Hour), "minute", true
	case "7d":
		return now.Add(-7 * 24 * time.Hour), "hour", true
	default:
		return time.Time{}, "", false
	}
}
