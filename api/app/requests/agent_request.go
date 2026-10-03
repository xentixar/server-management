package requests

import "api/app/models"

type AgentHeartbeatRequest struct {
	ProtocolVersion int                        `json:"protocol_version"`
	AgentVersion    string                     `json:"agent_version"`
	Hostname        string                     `json:"hostname"`
	Metrics         *models.HostMetricsPayload `json:"metrics"`
}

type AgentJobResultRequest struct {
	OK       bool           `json:"ok"`
	Message  string         `json:"message"`
	ExitCode int            `json:"exit_code"`
	Secrets  map[string]any `json:"secrets"`
}
