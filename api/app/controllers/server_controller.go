package controllers

import (
	"errors"
	"strings"

	"api/app/constants"
	"api/app/models"
	"api/app/policies"
	"api/app/provision"
	"api/app/requests"
	"api/app/resources"
	"api/config"

	"github.com/gin-gonic/gin"
)

type ServerController struct{}

func (ServerController) List(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}

	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.ServerPolicy{}).View(member)) {
		return
	}

	servers, err := (models.Server{}).FindByOrganization(ctx, organizationID)
	if err != nil {
		respond(ctx).fail("Failed to fetch servers", gin.H{}, 500)
		return
	}

	respond(ctx).ok(
		"Servers fetched successfully!",
		resources.ServerCollection(servers),
		200,
	)
}

func (ServerController) TestConnection(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}

	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.ServerPolicy{}).Create(member)) {
		return
	}

	var req requests.TestConnectionRequest
	if !bindJSON(ctx, &req) {
		return
	}
	if !validateServerAuth(ctx, req.AuthenticationMethod, req.PrivateKey, req.Password, true) {
		return
	}

	result := provision.TestConnection(sshTarget(req.ServerSSHFields))
	if !result.OK {
		respond(ctx).fail("SSH connection failed", gin.H{
			"ok":      false,
			"message": result.Message,
		}, 422)
		return
	}

	respond(ctx).ok("SSH connection successful", gin.H{
		"ok":      true,
		"message": result.Message,
	}, 200)
}

func (ServerController) Create(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}

	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.ServerPolicy{}).Create(member)) {
		return
	}

	var req requests.CreateServerRequest
	if !bindJSON(ctx, &req) {
		return
	}
	if !validateServerAuth(ctx, req.AuthenticationMethod, req.PrivateKey, req.Password, true) {
		return
	}

	target := sshTarget(req.ServerSSHFields)

	if test := provision.TestConnection(target); !test.OK {
		respond(ctx).fail("SSH connection failed", gin.H{
			"ok":      false,
			"message": test.Message,
		}, 422)
		return
	}

	server := models.Server{
		OrganizationID:       organizationID,
		Hostname:             strings.TrimSpace(req.Hostname),
		IPAddress:            strings.TrimSpace(req.IPAddress),
		SSHPort:              req.SSHPort,
		SSHUser:              strings.TrimSpace(req.SSHUser),
		AuthenticationMethod: req.AuthenticationMethod,
		PrivateKey:           req.PrivateKey,
		Passphrase:           req.Passphrase,
		Password:             req.Password,
		Region:               strings.TrimSpace(req.Region),
		Labels:               req.Labels,
		Notes:                req.Notes,
		AgentStatus:          constants.AgentStatusPending,
	}

	if err := server.Create(ctx); err != nil {
		respond(ctx).fail("Failed to create server", gin.H{}, 500)
		return
	}

	boot := provision.InstallAgent(target, config.PublicAPIURL(), server.AgentToken)
	bootstrap := gin.H{
		"ok":      boot.OK,
		"message": boot.Message,
	}

	respond(ctx).ok(
		"Server created successfully!",
		resources.ServerCreatedResource(server, bootstrap),
		201,
	)
}

func (ServerController) Bootstrap(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}
	serverID, ok := uintParam(ctx, "server", "Invalid server ID")
	if !ok {
		return
	}

	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.ServerPolicy{}).Update(member)) {
		return
	}

	server, err := (models.Server{}).FindByID(ctx, serverID, organizationID)
	if err != nil {
		respond(ctx).fail("Server not found", gin.H{}, 404)
		return
	}

	target := provision.TargetFromServer(
		server.IPAddress,
		server.SSHUser,
		server.AuthenticationMethod,
		server.PrivateKey,
		server.Password,
		server.Passphrase,
		server.SSHPort,
	)
	boot := provision.InstallAgent(target, config.PublicAPIURL(), server.AgentToken)
	if !boot.OK {
		ctx.Error(errors.New("bootstrap: " + boot.Message))
		respond(ctx).fail("Agent install failed", gin.H{
			"ok":      false,
			"message": boot.Message,
		}, 422)
		return
	}

	respond(ctx).ok("Agent installed", gin.H{
		"ok":      true,
		"message": boot.Message,
	}, 200)
}

func (ServerController) Get(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}
	serverID, ok := uintParam(ctx, "server", "Invalid server ID")
	if !ok {
		return
	}

	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.ServerPolicy{}).View(member)) {
		return
	}

	server, err := (models.Server{}).FindByID(ctx, serverID, organizationID)
	if err != nil {
		respond(ctx).fail("Server not found", gin.H{}, 404)
		return
	}

	respond(ctx).ok(
		"Server fetched successfully!",
		resources.ServerResource(server),
		200,
	)
}

func (ServerController) Update(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}
	serverID, ok := uintParam(ctx, "server", "Invalid server ID")
	if !ok {
		return
	}

	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.ServerPolicy{}).Update(member)) {
		return
	}

	server, err := (models.Server{}).FindByID(ctx, serverID, organizationID)
	if err != nil {
		respond(ctx).fail("Server not found", gin.H{}, 404)
		return
	}

	var req requests.UpdateServerRequest
	if !bindJSON(ctx, &req) {
		return
	}
	if !validateServerAuth(ctx, req.AuthenticationMethod, req.PrivateKey, req.Password, false) {
		return
	}

	server.Hostname = strings.TrimSpace(req.Hostname)
	server.IPAddress = strings.TrimSpace(req.IPAddress)
	server.SSHPort = req.SSHPort
	server.SSHUser = strings.TrimSpace(req.SSHUser)
	server.AuthenticationMethod = req.AuthenticationMethod
	server.Region = strings.TrimSpace(req.Region)
	server.Labels = req.Labels
	server.Notes = req.Notes

	if req.PrivateKey != "" {
		server.PrivateKey = req.PrivateKey
	}
	if req.Passphrase != "" {
		server.Passphrase = req.Passphrase
	}
	if req.Password != "" {
		server.Password = req.Password
	}

	if err := server.Update(ctx); err != nil {
		respond(ctx).fail("Failed to update server", gin.H{}, 500)
		return
	}

	respond(ctx).ok(
		"Server updated successfully!",
		resources.ServerResource(server),
		200,
	)
}

func (ServerController) Delete(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}
	serverID, ok := uintParam(ctx, "server", "Invalid server ID")
	if !ok {
		return
	}

	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.ServerPolicy{}).Delete(member)) {
		return
	}

	server, err := (models.Server{}).FindByID(ctx, serverID, organizationID)
	if err != nil {
		respond(ctx).fail("Server not found", gin.H{}, 404)
		return
	}

	target := provision.TargetFromServer(
		server.IPAddress,
		server.SSHUser,
		server.AuthenticationMethod,
		server.PrivateKey,
		server.Password,
		server.Passphrase,
		server.SSHPort,
	)
	cleanup := provision.UninstallAgent(target)
	_ = (models.AgentJob{}).DeleteForServer(ctx, server.ID)
	_ = (models.ServerService{}).DeleteForServer(ctx, server.ID)
	_ = (models.ServerCron{}).DeleteForServer(ctx, server.ID)

	if err := server.Delete(ctx); err != nil {
		respond(ctx).fail("Failed to delete server", gin.H{}, 500)
		return
	}

	msg := "Server deleted"
	if cleanup.OK {
		msg = "Server deleted and agent removed from host"
	} else {
		msg = "Server deleted (agent cleanup failed: " + cleanup.Message + ")"
	}
	respond(ctx).ok(msg, gin.H{
		"agent_removed": cleanup.OK,
		"cleanup":       cleanup.Message,
	}, 200)
}

func sshTarget(fields requests.ServerSSHFields) provision.Target {
	return provision.TargetFromServer(
		strings.TrimSpace(fields.IPAddress),
		strings.TrimSpace(fields.SSHUser),
		fields.AuthenticationMethod,
		fields.PrivateKey,
		fields.Password,
		fields.Passphrase,
		fields.SSHPort,
	)
}

func validateServerAuth(ctx *gin.Context, method, privateKey, password string, requireSecret bool) bool {
	switch method {
	case constants.AuthMethodKey:
		if requireSecret && strings.TrimSpace(privateKey) == "" {
			respond(ctx).fail("Validation error", gin.H{"private_key": "Private key is required"}, 422)
			return false
		}
	case constants.AuthMethodPassword:
		if requireSecret && strings.TrimSpace(password) == "" {
			respond(ctx).fail("Validation error", gin.H{"password": "Password is required"}, 422)
			return false
		}
	default:
		respond(ctx).fail("Validation error", gin.H{"auth_method": "Invalid authentication method"}, 422)
		return false
	}
	return true
}
