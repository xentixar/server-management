package constants

const (
	AuthMethodKey      = "key"
	AuthMethodPassword = "password"

	AgentStatusPending  = "pending"
	AgentStatusOnline   = "online"
	AgentStatusOffline  = "offline"
	AgentStatusDegraded = "degraded"
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

func IsValidAuthMethod(method string) bool {
	_, ok := authMethods[method]
	return ok
}

func IsValidAgentStatus(status string) bool {
	_, ok := agentStatuses[status]
	return ok
}
