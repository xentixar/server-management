package controllers

import (
	"net/http"
	"strconv"

	"api/app/constants"
	"api/app/models"
	"api/app/requests"

	"github.com/gin-gonic/gin"
)

type AgentController struct{}

func currentServer(ctx *gin.Context) models.Server {
	return ctx.MustGet("server").(models.Server)
}

func (AgentController) Heartbeat(ctx *gin.Context) {
	var req requests.AgentHeartbeatRequest
	if !bindJSON(ctx, &req) {
		return
	}

	server := currentServer(ctx)
	if err := server.RecordHeartbeat(ctx, req.AgentVersion); err != nil {
		respond(ctx).fail("Failed to record heartbeat", gin.H{}, http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"ok":               true,
		"protocol_version": constants.AgentProtocolVersion,
	})
}

func (AgentController) NextJob(ctx *gin.Context) {
	server := currentServer(ctx)

	job, err := (models.AgentJob{}).ClaimNext(ctx, server.ID)
	if err != nil {
		respond(ctx).fail("Failed to claim job", gin.H{}, http.StatusInternalServerError)
		return
	}
	if job == nil {
		ctx.Status(http.StatusNoContent)
		return
	}

	payload := job.Payload
	if payload == nil {
		payload = map[string]any{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":      job.PublicID(),
		"kind":    job.Kind,
		"payload": payload,
	})
}

func (AgentController) ReportJob(ctx *gin.Context) {
	jobID, err := strconv.ParseUint(ctx.Param("job"), 10, 64)
	if err != nil || jobID == 0 {
		respond(ctx).fail("Invalid job id", gin.H{}, http.StatusBadRequest)
		return
	}

	var req requests.AgentJobResultRequest
	if !bindJSON(ctx, &req) {
		return
	}

	server := currentServer(ctx)
	job, err := (models.AgentJob{}).FindForServer(ctx, uint(jobID), server.ID)
	if err != nil {
		respond(ctx).fail("Job not found", gin.H{}, http.StatusNotFound)
		return
	}

	if err := job.Complete(ctx, req.OK, req.Message, req.ExitCode); err != nil {
		respond(ctx).fail(err.Error(), gin.H{}, http.StatusConflict)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
