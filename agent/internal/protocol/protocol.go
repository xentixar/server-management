// Package protocol defines the agent ↔ control-plane wire types and version.
package protocol

// Version is negotiated on heartbeat. Bump when request/response shapes break.
const Version = 1

// AgentVersion is the binary release string (set via -ldflags in releases).
var AgentVersion = "dev"

type HeartbeatRequest struct {
	ProtocolVersion int    `json:"protocol_version"`
	AgentVersion    string `json:"agent_version"`
	Hostname        string `json:"hostname,omitempty"`
}

type HeartbeatResponse struct {
	OK              bool `json:"ok"`
	ProtocolVersion int  `json:"protocol_version"`
}

type Job struct {
	ID      string         `json:"id"`
	Kind    string         `json:"kind"`
	Payload map[string]any `json:"payload,omitempty"`
}

type JobResultRequest struct {
	OK       bool   `json:"ok"`
	Message  string `json:"message,omitempty"`
	ExitCode int    `json:"exit_code,omitempty"`
}
