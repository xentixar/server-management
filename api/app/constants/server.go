package constants

const (
	AuthMethodKey      = "key"
	AuthMethodPassword = "password"

	AgentStatusPending  = "pending"
	AgentStatusOnline   = "online"
	AgentStatusOffline  = "offline"
	AgentStatusDegraded = "degraded"

	AgentOnlineWithin   = 45  // seconds
	AgentDegradedWithin = 120 // seconds

	ProvisionStatusNone    = "none"
	ProvisionStatusPending = "pending"
	ProvisionStatusRunning = "running"
	ProvisionStatusReady   = "ready"
	ProvisionStatusFailed  = "failed"

	RuntimeStatusInstalled  = "installed"
	RuntimeStatusInstalling = "installing"
	RuntimeStatusFailed     = "failed"

	RuntimePHP  = "php"
	RuntimeNode = "node"

	AddonRedis    = "redis"
	AddonMySQL    = "mysql"
	AddonMariaDB  = "mariadb"
	AddonPostgres = "postgres"

	ServiceStatusPending = "pending"
	ServiceStatusRunning = "running"
	ServiceStatusStopped = "stopped"
	ServiceStatusFailed  = "failed"
	ServiceStatusUnknown = "unknown"

	ServiceControlStart   = "start"
	ServiceControlStop    = "stop"
	ServiceControlRestart = "restart"
	ServiceControlStatus  = "status"

	CronStatusPending = "pending"
	CronStatusSynced  = "synced"
	CronStatusFailed  = "failed"

	// DefaultProcessUser is the managed non-root Unix account created at provision (not a product brand).
	DefaultProcessUser = "sm"
	DefaultCronTZ      = "UTC"
)

var authMethods = map[string]struct{}{
	AuthMethodKey:      {},
	AuthMethodPassword: {},
}

var agentStatuses = map[string]struct{}{
	AgentStatusPending:  {},
	AgentStatusOnline:   {},
	AgentStatusOffline:  {},
	AgentStatusDegraded: {},
}

var provisionStatuses = map[string]struct{}{
	ProvisionStatusNone:    {},
	ProvisionStatusPending: {},
	ProvisionStatusRunning: {},
	ProvisionStatusReady:   {},
	ProvisionStatusFailed:  {},
}

// PHPCatalog and NodeCatalog are the installable versions for v1.
var PHPCatalog = []string{"8.1", "8.2", "8.3", "8.4", "8.5"}
var NodeCatalog = []string{"18", "20", "22", "24"}

func IsValidAuthMethod(method string) bool {
	_, ok := authMethods[method]
	return ok
}

func IsValidAgentStatus(status string) bool {
	_, ok := agentStatuses[status]
	return ok
}

func IsValidProvisionStatus(status string) bool {
	_, ok := provisionStatuses[status]
	return ok
}

func IsAllowedPHPVersion(version string) bool {
	for _, v := range PHPCatalog {
		if v == version {
			return true
		}
	}
	return false
}

func IsAllowedNodeVersion(version string) bool {
	for _, v := range NodeCatalog {
		if v == version {
			return true
		}
	}
	return false
}

var serviceStatuses = map[string]struct{}{
	ServiceStatusPending: {},
	ServiceStatusRunning: {},
	ServiceStatusStopped: {},
	ServiceStatusFailed:  {},
	ServiceStatusUnknown: {},
}

var serviceControls = map[string]struct{}{
	ServiceControlStart:   {},
	ServiceControlStop:    {},
	ServiceControlRestart: {},
	ServiceControlStatus:  {},
}

var cronStatuses = map[string]struct{}{
	CronStatusPending: {},
	CronStatusSynced:  {},
	CronStatusFailed:  {},
}

func IsValidServiceStatus(status string) bool {
	_, ok := serviceStatuses[status]
	return ok
}

func IsValidServiceControl(action string) bool {
	_, ok := serviceControls[action]
	return ok
}

func IsValidCronStatus(status string) bool {
	_, ok := cronStatuses[status]
	return ok
}
