package controllers

import (
	"net/http"
	"strconv"

	"api/app/constants"
	"api/app/models"
	"api/app/policies"
	"api/app/requests"
	"api/app/resources"

	"github.com/gin-gonic/gin"
)

func (ServerController) Provision(ctx *gin.Context) {
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
		respond(ctx).fail("Server not found", gin.H{}, http.StatusNotFound)
		return
	}
	if server.AgentStatus != constants.AgentStatusOnline {
		respond(ctx).fail("Agent must be online to provision", gin.H{}, http.StatusUnprocessableEntity)
		return
	}

	var req requests.ProvisionServerRequest
	if !bindJSON(ctx, &req) {
		return
	}
	if req.Addons.MySQL && req.Addons.MariaDB {
		respond(ctx).fail("Validation error", gin.H{
			"addons": "mysql and mariadb cannot both be selected",
		}, http.StatusUnprocessableEntity)
		return
	}

	phpVersions, errMsg := models.NormalizeVersionList(req.PHPVersions, constants.IsAllowedPHPVersion)
	if errMsg != "" {
		respond(ctx).fail("Validation error", gin.H{"php_versions": errMsg}, http.StatusUnprocessableEntity)
		return
	}
	nodeVersions, errMsg := models.NormalizeVersionList(req.NodeVersions, constants.IsAllowedNodeVersion)
	if errMsg != "" {
		respond(ctx).fail("Validation error", gin.H{"node_versions": errMsg}, http.StatusUnprocessableEntity)
		return
	}

	open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, constants.JobKindServerProvision)
	if err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	}
	if open {
		respond(ctx).fail("A provision job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	addonFlags := map[string]bool{
		constants.AddonRedis:    req.Addons.Redis,
		constants.AddonMySQL:    req.Addons.MySQL,
		constants.AddonMariaDB:  req.Addons.MariaDB,
		constants.AddonPostgres: req.Addons.Postgres,
	}
	addons := models.AddonList(addonFlags)

	payload := map[string]any{
		"php_versions":  phpVersions,
		"node_versions": nodeVersions,
		"addons": map[string]any{
			constants.AddonRedis:    req.Addons.Redis,
			constants.AddonMySQL:    req.Addons.MySQL,
			constants.AddonMariaDB:  req.Addons.MariaDB,
			constants.AddonPostgres: req.Addons.Postgres,
		},
	}

	server.Runtimes = models.SeedInstalling(phpVersions, nodeVersions, addons)
	server.ProvisionStatus = constants.ProvisionStatusPending
	if err := server.SaveProvision(ctx); err != nil {
		respond(ctx).fail("Failed to update server", gin.H{}, http.StatusInternalServerError)
		return
	}
	_ = server.ClearProvisionSecrets(ctx)

	job, err := (models.AgentJob{}).Enqueue(ctx, server.ID, constants.JobKindServerProvision, payload)
	if err != nil {
		respond(ctx).fail("Failed to enqueue provision job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Provision job queued", resources.AgentJobResource(job), http.StatusAccepted)
}

func (ServerController) GetRuntimes(ctx *gin.Context) {
	server, ok := loadServerForView(ctx)
	if !ok {
		return
	}
	if !requireProvisionReady(ctx, server) {
		return
	}
	server.Runtimes.Normalize()
	respond(ctx).ok("Runtimes fetched", server.Runtimes, http.StatusOK)
}

func (ServerController) RuntimesCatalog(ctx *gin.Context) {
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

	respond(ctx).ok("Runtime catalog", gin.H{
		"php":  constants.PHPCatalog,
		"node": constants.NodeCatalog,
	}, http.StatusOK)
}

func (ServerController) InstallRuntime(ctx *gin.Context) {
	server, ok := loadServerForUpdate(ctx)
	if !ok {
		return
	}
	if server.AgentStatus != constants.AgentStatusOnline {
		respond(ctx).fail("Agent must be online", gin.H{}, http.StatusUnprocessableEntity)
		return
	}
	if !requireProvisionReady(ctx, server) {
		return
	}

	var req requests.InstallRuntimeRequest
	if !bindJSON(ctx, &req) {
		return
	}

	allow := constants.IsAllowedPHPVersion
	kind := constants.JobKindRuntimePHPInstall
	if req.Runtime == constants.RuntimeNode {
		allow = constants.IsAllowedNodeVersion
		kind = constants.JobKindRuntimeNodeInstall
	}
	if !allow(req.Version) {
		respond(ctx).fail("Validation error", gin.H{"version": "unsupported version"}, http.StatusUnprocessableEntity)
		return
	}

	open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, kind)
	if err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	}
	if open {
		respond(ctx).fail("A runtime install job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	family := server.Runtimes.Family(req.Runtime)
	if family == nil {
		respond(ctx).fail("Invalid runtime", gin.H{}, http.StatusBadRequest)
		return
	}
	if v := family.Find(req.Version); v != nil && v.Status == constants.RuntimeStatusInstalled {
		respond(ctx).fail("Version already installed", gin.H{}, http.StatusConflict)
		return
	}

	setDefault := req.SetDefault || family.Default == ""
	family.Upsert(req.Version, constants.RuntimeStatusInstalling)
	if err := server.SaveProvision(ctx); err != nil {
		respond(ctx).fail("Failed to update runtimes", gin.H{}, http.StatusInternalServerError)
		return
	}

	job, err := (models.AgentJob{}).Enqueue(ctx, server.ID, kind, map[string]any{
		"runtime":     req.Runtime,
		"version":     req.Version,
		"set_default": setDefault,
	})
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Runtime install queued", resources.AgentJobResource(job), http.StatusAccepted)
}

func (ServerController) RemoveRuntime(ctx *gin.Context) {
	server, ok := loadServerForUpdate(ctx)
	if !ok {
		return
	}
	if !requireProvisionReady(ctx, server) {
		return
	}
	// ponytail: remove jobs not implemented on agent yet
	respond(ctx).fail("Runtime removal is not available yet", gin.H{}, http.StatusNotImplemented)
}

func (ServerController) SetRuntimeDefault(ctx *gin.Context) {
	server, ok := loadServerForUpdate(ctx)
	if !ok {
		return
	}
	if server.AgentStatus != constants.AgentStatusOnline {
		respond(ctx).fail("Agent must be online", gin.H{}, http.StatusUnprocessableEntity)
		return
	}
	if !requireProvisionReady(ctx, server) {
		return
	}

	var req requests.SetRuntimeDefaultRequest
	if !bindJSON(ctx, &req) {
		return
	}

	family := server.Runtimes.Family(req.Runtime)
	if family == nil {
		respond(ctx).fail("Invalid runtime", gin.H{}, http.StatusBadRequest)
		return
	}
	v := family.Find(req.Version)
	if v == nil || v.Status != constants.RuntimeStatusInstalled {
		respond(ctx).fail("Version is not installed", gin.H{}, http.StatusUnprocessableEntity)
		return
	}
	kind := constants.JobKindRuntimePHPDefault
	if req.Runtime == constants.RuntimeNode {
		kind = constants.JobKindRuntimeNodeDefault
	}
	open, err := (models.AgentJob{}).HasOpen(ctx, server.ID, kind)
	if err != nil {
		respond(ctx).fail("Failed to check jobs", gin.H{}, http.StatusInternalServerError)
		return
	}
	if open {
		respond(ctx).fail("A set-default job is already in progress", gin.H{}, http.StatusConflict)
		return
	}

	job, err := (models.AgentJob{}).Enqueue(ctx, server.ID, kind, map[string]any{
		"runtime": req.Runtime,
		"version": req.Version,
	})
	if err != nil {
		respond(ctx).fail("Failed to enqueue job", gin.H{}, http.StatusInternalServerError)
		return
	}

	respond(ctx).ok("Set default queued", resources.AgentJobResource(job), http.StatusAccepted)
}

func (ServerController) GetProvisionSecrets(ctx *gin.Context) {
	server, ok := loadServerForView(ctx)
	if !ok {
		return
	}
	if !server.HasProvisionSecrets() {
		respond(ctx).fail("No provision secrets available", gin.H{}, http.StatusNotFound)
		return
	}
	respond(ctx).ok("Provision secrets", server.ProvisionSecrets, http.StatusOK)
}

func (ServerController) DismissProvisionSecrets(ctx *gin.Context) {
	server, ok := loadServerForUpdate(ctx)
	if !ok {
		return
	}
	if err := server.ClearProvisionSecrets(ctx); err != nil {
		respond(ctx).fail("Failed to dismiss secrets", gin.H{}, http.StatusInternalServerError)
		return
	}
	respond(ctx).ok("Provision secrets dismissed", gin.H{}, http.StatusOK)
}

func (ServerController) GetLatestJob(ctx *gin.Context) {
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

	if _, err := (models.Server{}).FindByID(ctx, serverID, organizationID); err != nil {
		respond(ctx).fail("Server not found", gin.H{}, http.StatusNotFound)
		return
	}

	kind := ctx.Query("kind")
	job, err := (models.AgentJob{}).LatestByKind(ctx, serverID, kind)
	if err != nil {
		respond(ctx).fail("Failed to fetch job", gin.H{}, http.StatusInternalServerError)
		return
	}
	if job == nil {
		respond(ctx).fail("Job not found", gin.H{}, http.StatusNotFound)
		return
	}

	respond(ctx).ok("Job fetched", resources.AgentJobResource(*job), http.StatusOK)
}

func (ServerController) GetJob(ctx *gin.Context) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return
	}
	serverID, ok := uintParam(ctx, "server", "Invalid server ID")
	if !ok {
		return
	}
	jobID, err := strconv.ParseUint(ctx.Param("job"), 10, 64)
	if err != nil || jobID == 0 {
		respond(ctx).fail("Invalid job id", gin.H{}, http.StatusBadRequest)
		return
	}

	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return
	}
	if !denyUnless(ctx, (policies.ServerPolicy{}).View(member)) {
		return
	}

	if _, err := (models.Server{}).FindByID(ctx, serverID, organizationID); err != nil {
		respond(ctx).fail("Server not found", gin.H{}, http.StatusNotFound)
		return
	}

	job, err := (models.AgentJob{}).FindForServer(ctx, uint(jobID), serverID)
	if err != nil {
		respond(ctx).fail("Job not found", gin.H{}, http.StatusNotFound)
		return
	}

	respond(ctx).ok("Job fetched", resources.AgentJobResource(job), http.StatusOK)
}

func requireProvisionReady(ctx *gin.Context, server models.Server) bool {
	if server.ProvisionStatus == constants.ProvisionStatusReady {
		return true
	}
	respond(ctx).fail("Provision the server stack first", gin.H{}, http.StatusUnprocessableEntity)
	return false
}

func loadServerForView(ctx *gin.Context) (models.Server, bool) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return models.Server{}, false
	}
	serverID, ok := uintParam(ctx, "server", "Invalid server ID")
	if !ok {
		return models.Server{}, false
	}
	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return models.Server{}, false
	}
	if !denyUnless(ctx, (policies.ServerPolicy{}).View(member)) {
		return models.Server{}, false
	}
	server, err := (models.Server{}).FindByID(ctx, serverID, organizationID)
	if err != nil {
		respond(ctx).fail("Server not found", gin.H{}, http.StatusNotFound)
		return models.Server{}, false
	}
	return server, true
}

func loadServerForUpdate(ctx *gin.Context) (models.Server, bool) {
	organizationID, ok := uintParam(ctx, "organization", "Invalid organization ID")
	if !ok {
		return models.Server{}, false
	}
	serverID, ok := uintParam(ctx, "server", "Invalid server ID")
	if !ok {
		return models.Server{}, false
	}
	_, member, ok := requireOrgMember(ctx, organizationID)
	if !ok {
		return models.Server{}, false
	}
	if !denyUnless(ctx, (policies.ServerPolicy{}).Update(member)) {
		return models.Server{}, false
	}
	server, err := (models.Server{}).FindByID(ctx, serverID, organizationID)
	if err != nil {
		respond(ctx).fail("Server not found", gin.H{}, http.StatusNotFound)
		return models.Server{}, false
	}
	return server, true
}
