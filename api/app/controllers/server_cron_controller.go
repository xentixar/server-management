package controllers

import (
	"errors"
	"net/http"
	"strings"

	"api/app/constants"
	"api/app/models"
	"api/app/requests"
	"api/app/resources"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (ServerController) ListCrons(ctx *gin.Context) {
	server, ok := loadServerForCronsView(ctx)
	if !ok {
		return
	}
	if !requireProvisionReady(ctx, server) {
		return
	}
	items, err := (models.ServerCron{}).ListByServer(ctx, server.ID)
	if err != nil {
		respond(ctx).fail("Failed to list crons", gin.H{}, http.StatusInternalServerError)
		return
	}
	respond(ctx).ok("Crons fetched", resources.Collect(items, resources.ServerCronResource), http.StatusOK)
}

func (ServerController) CreateCron(ctx *gin.Context) {
	server, ok := loadServerForCronsManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	var req requests.CreateServerCronRequest
	if !bindJSON(ctx, &req) {
		return
	}

	cron := models.ServerCron{
		ServerID: server.ID,
		Schedule: req.Schedule,
		Command:  req.Command,
		User:     req.User,
		Timezone: req.Timezone,
		Enabled:  true,
		Status:   constants.CronStatusPending,
	}
	if strings.TrimSpace(cron.User) == "" {
		cron.User = server.ProcessUser()
	}
	if req.Enabled != nil {
		cron.Enabled = *req.Enabled
	}
	if err := cron.Validate(); err != nil {
		respond(ctx).fail("Validation error", gin.H{"error": err.Error()}, http.StatusUnprocessableEntity)
		return
	}

	if open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindCronSync); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("A cron sync job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	if err := cron.Create(ctx); err != nil {
		respond(ctx).fail("Failed to create cron", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := enqueueCronSync(ctx, server.ID)
	if err != nil {
		_ = cron.Delete(ctx)
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Cron create queued", gin.H{
		"job":  resources.AgentJobResource(job),
		"cron": resources.ServerCronResource(cron),
	}, http.StatusAccepted)
}

func (ServerController) UpdateCron(ctx *gin.Context) {
	server, ok := loadServerForCronsManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	cron, ok := loadCron(ctx, server.ID)
	if !ok {
		return
	}

	var req requests.UpdateServerCronRequest
	if !bindJSON(ctx, &req) {
		return
	}
	if req.Schedule != nil {
		cron.Schedule = *req.Schedule
	}
	if req.Command != nil {
		cron.Command = *req.Command
	}
	if req.User != nil {
		cron.User = *req.User
	}
	if req.Timezone != nil {
		cron.Timezone = *req.Timezone
	}
	if req.Enabled != nil {
		cron.Enabled = *req.Enabled
	}
	cron.Status = constants.CronStatusPending
	cron.Message = ""
	if err := cron.Validate(); err != nil {
		respond(ctx).fail("Validation error", gin.H{"error": err.Error()}, http.StatusUnprocessableEntity)
		return
	}

	if open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindCronSync); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("A cron sync job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	if err := cron.Save(ctx); err != nil {
		respond(ctx).fail("Failed to update cron", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := enqueueCronSync(ctx, server.ID)
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Cron update queued", gin.H{
		"job":  resources.AgentJobResource(job),
		"cron": resources.ServerCronResource(cron),
	}, http.StatusAccepted)
}

func (ServerController) FetchCronLogs(ctx *gin.Context) {
	server, ok := loadServerForCronsView(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	cron, ok := loadCron(ctx, server.ID)
	if !ok {
		return
	}

	if open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindCronLogs); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("A cron logs job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	job, err := (models.AgentJob{}).Enqueue(ctx, server.ID, constants.JobKindCronLogs, map[string]any{
		"uid": cron.PublicID(),
	})
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Cron logs queued", gin.H{
		"job":  resources.AgentJobResource(job),
		"cron": resources.ServerCronResource(cron),
	}, http.StatusAccepted)
}

func (ServerController) DeleteCron(ctx *gin.Context) {
	server, ok := loadServerForCronsManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	cron, ok := loadCron(ctx, server.ID)
	if !ok {
		return
	}

	if open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindCronSync); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("A cron sync job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	if err := cron.Delete(ctx); err != nil {
		respond(ctx).fail("Failed to delete cron", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := enqueueCronSync(ctx, server.ID)
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Cron delete queued", gin.H{
		"job": resources.AgentJobResource(job),
	}, http.StatusAccepted)
}

func loadCron(ctx *gin.Context, serverID uint) (models.ServerCron, bool) {
	id := strings.TrimSpace(ctx.Param("cron"))
	if id == "" {
		respond(ctx).fail("Invalid cron id", gin.H{}, http.StatusBadRequest)
		return models.ServerCron{}, false
	}
	cron, err := (models.ServerCron{}).FindByPublicID(ctx, id, serverID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respond(ctx).fail("Cron not found", gin.H{}, http.StatusNotFound)
		return models.ServerCron{}, false
	}
	if err != nil {
		respond(ctx).fail("Failed to load cron", gin.H{}, http.StatusInternalServerError)
		return models.ServerCron{}, false
	}
	return cron, true
}

func enqueueCronSync(ctx *gin.Context, serverID uint) (models.AgentJob, error) {
	crons, err := (models.ServerCron{}).ListByServer(ctx, serverID)
	if err != nil {
		return models.AgentJob{}, err
	}
	for i := range crons {
		crons[i].Status = constants.CronStatusPending
		if err := crons[i].Save(ctx); err != nil {
			return models.AgentJob{}, err
		}
	}
	return (models.AgentJob{}).Enqueue(ctx, serverID, constants.JobKindCronSync, models.CronSyncPayload(crons))
}
