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

func (ServerController) ListServices(ctx *gin.Context) {
	server, ok := loadServerForServicesView(ctx)
	if !ok {
		return
	}
	if !requireProvisionReady(ctx, server) {
		return
	}
	items, err := (models.ServerService{}).ListByServer(ctx, server.ID)
	if err != nil {
		respond(ctx).fail("Failed to list services", gin.H{}, http.StatusInternalServerError)
		return
	}
	respond(ctx).ok("Services fetched", resources.Collect(items, resources.ServerServiceResource), http.StatusOK)
}

func (ServerController) CreateService(ctx *gin.Context) {
	server, ok := loadServerForServicesManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	var req requests.CreateServerServiceRequest
	if !bindJSON(ctx, &req) {
		return
	}

	svc := models.ServerService{
		ServerID:    server.ID,
		Name:        strings.ToLower(strings.TrimSpace(req.Name)),
		Command:     req.Command,
		Directory:   req.Directory,
		User:        req.User,
		Processes:   req.Processes,
		AutoRestart: true,
		Autostart:   true,
		Enabled:     true,
		Status:      constants.ServiceStatusPending,
	}
	if strings.TrimSpace(svc.User) == "" {
		svc.User = server.ProcessUser()
	}
	if req.AutoRestart != nil {
		svc.AutoRestart = *req.AutoRestart
	}
	if req.Autostart != nil {
		svc.Autostart = *req.Autostart
	}
	if req.Enabled != nil {
		svc.Enabled = *req.Enabled
	}
	if err := svc.Validate(); err != nil {
		respond(ctx).fail("Validation error", gin.H{"error": err.Error()}, http.StatusUnprocessableEntity)
		return
	}

	if open, err := serviceJobsOpen(ctx, server.ID); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("A service job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	if err := svc.Create(ctx); err != nil {
		respond(ctx).fail("Failed to create service", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := (models.AgentJob{}).Enqueue(ctx, server.ID, constants.JobKindServiceApply, svc.JobPayload())
	if err != nil {
		_ = svc.Delete(ctx)
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Service create queued", gin.H{
		"job":     resources.AgentJobResource(job),
		"service": resources.ServerServiceResource(svc),
	}, http.StatusAccepted)
}

func (ServerController) UpdateService(ctx *gin.Context) {
	server, ok := loadServerForServicesManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	svc, ok := loadService(ctx, server.ID)
	if !ok {
		return
	}

	var req requests.UpdateServerServiceRequest
	if !bindJSON(ctx, &req) {
		return
	}
	if req.Command != nil {
		svc.Command = *req.Command
	}
	if req.Directory != nil {
		svc.Directory = *req.Directory
	}
	if req.User != nil {
		svc.User = *req.User
	}
	if req.Processes != nil {
		svc.Processes = *req.Processes
	}
	if req.AutoRestart != nil {
		svc.AutoRestart = *req.AutoRestart
	}
	if req.Autostart != nil {
		svc.Autostart = *req.Autostart
	}
	if req.Enabled != nil {
		svc.Enabled = *req.Enabled
	}
	svc.Status = constants.ServiceStatusPending
	svc.Message = ""
	if err := svc.Validate(); err != nil {
		respond(ctx).fail("Validation error", gin.H{"error": err.Error()}, http.StatusUnprocessableEntity)
		return
	}

	if open, err := serviceJobsOpen(ctx, server.ID); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("A service job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	if err := svc.Save(ctx); err != nil {
		respond(ctx).fail("Failed to update service", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := (models.AgentJob{}).Enqueue(ctx, server.ID, constants.JobKindServiceApply, svc.JobPayload())
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Service update queued", gin.H{
		"job":     resources.AgentJobResource(job),
		"service": resources.ServerServiceResource(svc),
	}, http.StatusAccepted)
}

func (ServerController) DeleteService(ctx *gin.Context) {
	server, ok := loadServerForServicesManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	svc, ok := loadService(ctx, server.ID)
	if !ok {
		return
	}

	if open, err := serviceJobsOpen(ctx, server.ID); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("A service job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	svc.Status = constants.ServiceStatusPending
	svc.Message = "removing"
	if err := svc.Save(ctx); err != nil {
		respond(ctx).fail("Failed to update service", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := (models.AgentJob{}).Enqueue(ctx, server.ID, constants.JobKindServiceRemove, map[string]any{
		"uid":  svc.PublicID(),
		"name": svc.Name,
	})
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Service remove queued", gin.H{
		"job":     resources.AgentJobResource(job),
		"service": resources.ServerServiceResource(svc),
	}, http.StatusAccepted)
}

func (ServerController) ControlService(ctx *gin.Context) {
	var req requests.ControlServerServiceRequest
	if !bindJSON(ctx, &req) {
		return
	}
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if !constants.IsValidServiceControl(action) {
		respond(ctx).fail("Validation error", gin.H{"action": "must be start, stop, restart, or status"}, http.StatusUnprocessableEntity)
		return
	}

	// status/logs are view; start/stop/restart need manage
	var server models.Server
	var ok bool
	if action == constants.ServiceControlStatus {
		server, ok = loadServerForServicesView(ctx)
	} else {
		server, ok = loadServerForServicesManage(ctx)
	}
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	svc, ok := loadService(ctx, server.ID)
	if !ok {
		return
	}

	if open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindServiceControl); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("A service control job is already in progress", gin.H{}, http.StatusConflict)
		return
	}
	if open, err := (models.AgentJob{}).HasOpenKinds(ctx, server.ID, constants.JobKindServiceApply, constants.JobKindServiceRemove); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("A service apply/remove job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	svc.Status = constants.ServiceStatusPending
	if err := svc.Save(ctx); err != nil {
		respond(ctx).fail("Failed to update service", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := (models.AgentJob{}).Enqueue(ctx, server.ID, constants.JobKindServiceControl, map[string]any{
		"uid":    svc.PublicID(),
		"name":   svc.Name,
		"action": action,
	})
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Service control queued", gin.H{
		"job":     resources.AgentJobResource(job),
		"service": resources.ServerServiceResource(svc),
	}, http.StatusAccepted)
}

func loadService(ctx *gin.Context, serverID uint) (models.ServerService, bool) {
	id := strings.TrimSpace(ctx.Param("service"))
	if id == "" {
		respond(ctx).fail("Invalid service id", gin.H{}, http.StatusBadRequest)
		return models.ServerService{}, false
	}
	svc, err := (models.ServerService{}).FindByPublicID(ctx, id, serverID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respond(ctx).fail("Service not found", gin.H{}, http.StatusNotFound)
		return models.ServerService{}, false
	}
	if err != nil {
		respond(ctx).fail("Failed to load service", gin.H{}, http.StatusInternalServerError)
		return models.ServerService{}, false
	}
	return svc, true
}

func serviceJobsOpen(ctx *gin.Context, serverID uint) (bool, error) {
	return (models.AgentJob{}).HasOpenKinds(ctx, serverID,
		constants.JobKindServiceApply,
		constants.JobKindServiceRemove,
		constants.JobKindServiceControl,
	)
}

func requireAgentOnlineReady(ctx *gin.Context, server models.Server) bool {
	if server.EffectiveAgentStatus() != constants.AgentStatusOnline {
		respond(ctx).fail("Agent must be online", gin.H{}, http.StatusUnprocessableEntity)
		return false
	}
	return requireProvisionReady(ctx, server)
}
