// Package protocol defines the agent ↔ control-plane wire types and version.
package protocol

// Version is negotiated on heartbeat. Bump when request/response shapes break.
const Version = 1

// AgentVersion is the binary release string (set via -ldflags in releases).
var AgentVersion = "dev"

// HostMetrics is an optional heartbeat payload with host resource samples.
type HostMetrics struct {
	CPUPercent     *float64 `json:"cpu_percent,omitempty"`
	MemUsedBytes   *uint64  `json:"mem_used_bytes,omitempty"`
	MemTotalBytes  *uint64  `json:"mem_total_bytes,omitempty"`
	DiskUsedBytes  *uint64  `json:"disk_used_bytes,omitempty"`
	DiskTotalBytes *uint64  `json:"disk_total_bytes,omitempty"`
	Load1          *float64 `json:"load1,omitempty"`
	Load5          *float64 `json:"load5,omitempty"`
	Load15         *float64 `json:"load15,omitempty"`
}

type HeartbeatRequest struct {
	ProtocolVersion int          `json:"protocol_version"`
	AgentVersion    string       `json:"agent_version"`
	Hostname        string       `json:"hostname,omitempty"`
	Metrics         *HostMetrics `json:"metrics,omitempty"`
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
	OK       bool           `json:"ok"`
	Message  string         `json:"message,omitempty"`
	ExitCode int            `json:"exit_code,omitempty"`
	Secrets  map[string]any `json:"secrets,omitempty"`
}
