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

func (ServerController) ListSshKeys(ctx *gin.Context) {
	server, ok := loadServerForSshKeysView(ctx)
	if !ok {
		return
	}
	if !requireProvisionReady(ctx, server) {
		return
	}
	items, err := (models.ServerSshKey{}).ListByServer(ctx, server.ID)
	if err != nil {
		respond(ctx).fail("Failed to list SSH keys", gin.H{}, http.StatusInternalServerError)
		return
	}
	respond(ctx).ok("SSH keys fetched", resources.Collect(items, resources.ServerSshKeyResource), http.StatusOK)
}

func (ServerController) CreateSshKey(ctx *gin.Context) {
	server, ok := loadServerForSshKeysManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	var req requests.CreateServerSshKeyRequest
	if !bindJSON(ctx, &req) {
		return
	}

	key := models.ServerSshKey{
		ServerID:  server.ID,
		Name:      req.Name,
		PublicKey: req.PublicKey,
		Status:    constants.SshKeyStatusPending,
	}
	if err := key.Validate(); err != nil {
		respond(ctx).fail("Validation error", gin.H{"error": err.Error()}, http.StatusUnprocessableEntity)
		return
	}
	if dup, err := (models.ServerSshKey{}).FindByFingerprint(ctx, server.ID, key.Fingerprint, 0); err == nil && dup.ID != 0 {
		respond(ctx).fail("An SSH key with this fingerprint already exists", gin.H{}, http.StatusConflict)
		return
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		respond(ctx).fail("Failed to check SSH keys", gin.H{}, http.StatusInternalServerError)
		return
	}

	if open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindSSHKeySync); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("An SSH key sync job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	if err := key.Create(ctx); err != nil {
		respond(ctx).fail("Failed to create SSH key", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := enqueueSshKeySync(ctx, server)
	if err != nil {
		_ = key.Delete(ctx)
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("SSH key create queued", gin.H{
		"job": resources.AgentJobResource(job),
		"key": resources.ServerSshKeyResource(key),
	}, http.StatusAccepted)
}

func (ServerController) UpdateSshKey(ctx *gin.Context) {
	server, ok := loadServerForSshKeysManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	key, ok := loadSshKey(ctx, server.ID)
	if !ok {
		return
	}

	var req requests.UpdateServerSshKeyRequest
	if !bindJSON(ctx, &req) {
		return
	}
	if req.Name != nil {
		key.Name = *req.Name
	}
	if req.PublicKey != nil {
		key.PublicKey = *req.PublicKey
	}
	key.Status = constants.SshKeyStatusPending
	key.Message = ""
	if err := key.Validate(); err != nil {
		respond(ctx).fail("Validation error", gin.H{"error": err.Error()}, http.StatusUnprocessableEntity)
		return
	}
	if dup, err := (models.ServerSshKey{}).FindByFingerprint(ctx, server.ID, key.Fingerprint, key.ID); err == nil && dup.ID != 0 {
		respond(ctx).fail("An SSH key with this fingerprint already exists", gin.H{}, http.StatusConflict)
		return
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		respond(ctx).fail("Failed to check SSH keys", gin.H{}, http.StatusInternalServerError)
		return
	}

	if open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindSSHKeySync); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("An SSH key sync job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	if err := key.Save(ctx); err != nil {
		respond(ctx).fail("Failed to update SSH key", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := enqueueSshKeySync(ctx, server)
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("SSH key update queued", gin.H{
		"job": resources.AgentJobResource(job),
		"key": resources.ServerSshKeyResource(key),
	}, http.StatusAccepted)
}

func (ServerController) DeleteSshKey(ctx *gin.Context) {
	server, ok := loadServerForSshKeysManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	key, ok := loadSshKey(ctx, server.ID)
	if !ok {
		return
	}

	if open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindSSHKeySync); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("An SSH key sync job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	if err := key.Delete(ctx); err != nil {
		respond(ctx).fail("Failed to delete SSH key", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := enqueueSshKeySync(ctx, server)
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("SSH key delete queued", gin.H{
		"job": resources.AgentJobResource(job),
	}, http.StatusAccepted)
}

func (ServerController) SyncSshKeys(ctx *gin.Context) {
	server, ok := loadServerForSshKeysManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	if open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindSSHKeySync); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("An SSH key sync job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	job, err := enqueueSshKeySync(ctx, server)
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("SSH key sync queued", gin.H{
		"job": resources.AgentJobResource(job),
	}, http.StatusAccepted)
}

func loadSshKey(ctx *gin.Context, serverID uint) (models.ServerSshKey, bool) {
	id := strings.TrimSpace(ctx.Param("key"))
	if id == "" {
		respond(ctx).fail("Invalid SSH key id", gin.H{}, http.StatusBadRequest)
		return models.ServerSshKey{}, false
	}
	key, err := (models.ServerSshKey{}).FindByPublicID(ctx, id, serverID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respond(ctx).fail("SSH key not found", gin.H{}, http.StatusNotFound)
		return models.ServerSshKey{}, false
	}
	if err != nil {
		respond(ctx).fail("Failed to load SSH key", gin.H{}, http.StatusInternalServerError)
		return models.ServerSshKey{}, false
	}
	return key, true
}

func enqueueSshKeySync(ctx *gin.Context, server models.Server) (models.AgentJob, error) {
	keys, err := (models.ServerSshKey{}).ListByServer(ctx, server.ID)
	if err != nil {
		return models.AgentJob{}, err
	}
	for i := range keys {
		keys[i].Status = constants.SshKeyStatusPending
		if err := keys[i].Save(ctx); err != nil {
			return models.AgentJob{}, err
		}
	}
	return (models.AgentJob{}).Enqueue(
		ctx,
		server.ID,
		constants.JobKindSSHKeySync,
		models.SshKeySyncPayload(keys),
	)
}
