package resources

import (
	"api/app/models"
)

func ServerResource(data models.Server) map[string]any {
	labels := data.Labels
	if labels == nil {
		labels = []string{}
	}

	var lastSeenAt any
	if data.LastSeenAt != nil {
		lastSeenAt = data.LastSeenAt.UTC().Format("2006-01-02T15:04:05Z")
	}

	var provisionedAt any
	if data.ProvisionedAt != nil {
		provisionedAt = data.ProvisionedAt.UTC().Format("2006-01-02T15:04:05Z")
	}

	runtimes := data.Runtimes
	runtimes.Normalize()
	if data.ProvisionStatus == "" {
		data.ProvisionStatus = "none"
	}
	timezone := data.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	return map[string]any{
		"id":                    data.ID,
		"hostname":              data.Hostname,
		"ip":                    data.IPAddress,
		"ssh_port":              data.SSHPort,
		"ssh_user":              data.SSHUser,
		"auth_method":           data.AuthenticationMethod,
		"has_key":               data.PrivateKey != "",
		"region":                data.Region,
		"timezone":              timezone,
		"labels":                labels,
		"notes":                 data.Notes,
		"agent":                 data.EffectiveAgentStatus(),
		"agent_version":         data.AgentVersion,
		"last_seen_at":          lastSeenAt,
		"provision_status":      data.ProvisionStatus,
		"provisioned_at":        provisionedAt,
		"has_provision_secrets": data.HasProvisionSecrets(),
		"managed_user":          data.ProcessUser(),
		"runtimes":              runtimes,
		"sites":                 0,
		"created_at":            data.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func ServerCreatedResource(data models.Server, bootstrap map[string]any) map[string]any {
	out := ServerResource(data)
	if bootstrap != nil {
		out["bootstrap"] = bootstrap
	}
	return out
}

func ServerCollection(data []models.Server) []map[string]any {
	return Collect(data, ServerResource)
}
