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

func (ServerController) ListFirewallRules(ctx *gin.Context) {
	server, ok := loadServerForFirewallView(ctx)
	if !ok {
		return
	}
	if !requireProvisionReady(ctx, server) {
		return
	}
	items, err := (models.ServerFirewallRule{}).ListByServer(ctx, server.ID)
	if err != nil {
		respond(ctx).fail("Failed to list firewall rules", gin.H{}, http.StatusInternalServerError)
		return
	}
	respond(ctx).ok("Firewall rules fetched", resources.Collect(items, resources.ServerFirewallRuleResource), http.StatusOK)
}

func (ServerController) CreateFirewallRule(ctx *gin.Context) {
	server, ok := loadServerForFirewallManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	var req requests.CreateServerFirewallRuleRequest
	if !bindJSON(ctx, &req) {
		return
	}

	pos, err := (models.ServerFirewallRule{}).NextPosition(ctx, server.ID)
	if err != nil {
		respond(ctx).fail("Failed to allocate position", gin.H{}, http.StatusInternalServerError)
		return
	}
	if req.Position != nil {
		pos = *req.Position
	}

	rule := models.ServerFirewallRule{
		ServerID:  server.ID,
		Name:      req.Name,
		Action:    req.Action,
		Direction: req.Direction,
		Protocol:  req.Protocol,
		Source:    req.Source,
		Position:  pos,
		Enabled:   true,
		Status:    constants.FirewallStatusPending,
	}
	if req.PortStart != nil {
		rule.PortStart = *req.PortStart
	}
	if req.PortEnd != nil {
		rule.PortEnd = *req.PortEnd
	} else {
		rule.PortEnd = rule.PortStart
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if err := rule.Validate(); err != nil {
		respond(ctx).fail("Validation error", gin.H{"error": err.Error()}, http.StatusUnprocessableEntity)
		return
	}

	if open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindFirewallSync); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("A firewall sync job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	if err := rule.Create(ctx); err != nil {
		respond(ctx).fail("Failed to create firewall rule", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := enqueueFirewallSync(ctx, server)
	if err != nil {
		_ = rule.Delete(ctx)
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Firewall rule create queued", gin.H{
		"job":  resources.AgentJobResource(job),
		"rule": resources.ServerFirewallRuleResource(rule),
	}, http.StatusAccepted)
}

func (ServerController) UpdateFirewallRule(ctx *gin.Context) {
	server, ok := loadServerForFirewallManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	rule, ok := loadFirewallRule(ctx, server.ID)
	if !ok {
		return
	}

	var req requests.UpdateServerFirewallRuleRequest
	if !bindJSON(ctx, &req) {
		return
	}
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Action != nil {
		rule.Action = *req.Action
	}
	if req.Direction != nil {
		rule.Direction = *req.Direction
	}
	if req.Protocol != nil {
		rule.Protocol = *req.Protocol
	}
	if req.PortStart != nil {
		rule.PortStart = *req.PortStart
	}
	if req.PortEnd != nil {
		rule.PortEnd = *req.PortEnd
	}
	if req.Source != nil {
		rule.Source = *req.Source
	}
	if req.Position != nil {
		rule.Position = *req.Position
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	rule.Status = constants.FirewallStatusPending
	rule.Message = ""
	if err := rule.Validate(); err != nil {
		respond(ctx).fail("Validation error", gin.H{"error": err.Error()}, http.StatusUnprocessableEntity)
		return
	}

	if open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindFirewallSync); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("A firewall sync job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	if err := rule.Save(ctx); err != nil {
		respond(ctx).fail("Failed to update firewall rule", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := enqueueFirewallSync(ctx, server)
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Firewall rule update queued", gin.H{
		"job":  resources.AgentJobResource(job),
		"rule": resources.ServerFirewallRuleResource(rule),
	}, http.StatusAccepted)
}

func (ServerController) DeleteFirewallRule(ctx *gin.Context) {
	server, ok := loadServerForFirewallManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	rule, ok := loadFirewallRule(ctx, server.ID)
	if !ok {
		return
	}

	if open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindFirewallSync); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("A firewall sync job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	if err := rule.Delete(ctx); err != nil {
		respond(ctx).fail("Failed to delete firewall rule", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := enqueueFirewallSync(ctx, server)
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Firewall rule delete queued", gin.H{
		"job": resources.AgentJobResource(job),
	}, http.StatusAccepted)
}

func (ServerController) SyncFirewall(ctx *gin.Context) {
	server, ok := loadServerForFirewallManage(ctx)
	if !ok {
		return
	}
	if !requireAgentOnlineReady(ctx, server) {
		return
	}

	if open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindFirewallSync); err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	} else if open {
		respond(ctx).fail("A firewall sync job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	job, err := enqueueFirewallSync(ctx, server)
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Firewall sync queued", gin.H{
		"job": resources.AgentJobResource(job),
	}, http.StatusAccepted)
}

func loadFirewallRule(ctx *gin.Context, serverID uint) (models.ServerFirewallRule, bool) {
	id := strings.TrimSpace(ctx.Param("rule"))
	if id == "" {
		respond(ctx).fail("Invalid firewall rule id", gin.H{}, http.StatusBadRequest)
		return models.ServerFirewallRule{}, false
	}
	rule, err := (models.ServerFirewallRule{}).FindByPublicID(ctx, id, serverID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respond(ctx).fail("Firewall rule not found", gin.H{}, http.StatusNotFound)
		return models.ServerFirewallRule{}, false
	}
	if err != nil {
		respond(ctx).fail("Failed to load firewall rule", gin.H{}, http.StatusInternalServerError)
		return models.ServerFirewallRule{}, false
	}
	return rule, true
}

func enqueueFirewallSync(ctx *gin.Context, server models.Server) (models.AgentJob, error) {
	rules, err := (models.ServerFirewallRule{}).ListByServer(ctx, server.ID)
	if err != nil {
		return models.AgentJob{}, err
	}
	for i := range rules {
		rules[i].Status = constants.FirewallStatusPending
		if err := rules[i].Save(ctx); err != nil {
			return models.AgentJob{}, err
		}
	}
	return (models.AgentJob{}).Enqueue(
		ctx,
		server.ID,
		constants.JobKindFirewallSync,
		models.FirewallSyncPayload(server.SSHPort, rules),
	)
}
