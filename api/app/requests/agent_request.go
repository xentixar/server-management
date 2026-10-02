package requests

type AgentHeartbeatRequest struct {
	ProtocolVersion int    `json:"protocol_version"`
	AgentVersion    string `json:"agent_version"`
	Hostname        string `json:"hostname"`
}

type AgentJobResultRequest struct {
	OK       bool   `json:"ok"`
	Message  string `json:"message"`
	ExitCode int    `json:"exit_code"`
}
