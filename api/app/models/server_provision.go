package models

import (
	"context"
	"time"

	"api/app/constants"
)

func stringSliceFromPayload(payload map[string]any, key string) []string {
	raw, ok := payload[key]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func addonFlagsFromPayload(payload map[string]any) map[string]bool {
	flags := map[string]bool{}
	raw, ok := payload["addons"]
	if !ok || raw == nil {
		return flags
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return flags
	}
	for _, key := range []string{
		constants.AddonRedis,
		constants.AddonMySQL,
		constants.AddonMariaDB,
		constants.AddonPostgres,
	} {
		if b, ok := m[key].(bool); ok {
			flags[key] = b
		}
	}
	return flags
}

func (s *Server) ApplyJobClaim(ctx context.Context, job AgentJob) error {
	switch job.Kind {
	case constants.JobKindServerProvision:
		if s.ProvisionStatus == constants.ProvisionStatusPending {
			s.ProvisionStatus = constants.ProvisionStatusRunning
			return s.SaveProvision(ctx)
		}
	}
	return nil
}

func (s *Server) ApplyJobResult(ctx context.Context, job AgentJob, ok bool, secrets map[string]any) error {
	s.Runtimes.Normalize()

	switch job.Kind {
	case constants.JobKindServerProvision:
		php := stringSliceFromPayload(job.Payload, "php_versions")
		node := stringSliceFromPayload(job.Payload, "node_versions")
		addons := AddonList(addonFlagsFromPayload(job.Payload))
		status := constants.RuntimeStatusInstalled
		if !ok {
			status = constants.RuntimeStatusFailed
		}
		for _, v := range php {
			s.Runtimes.PHP.Upsert(v, status)
		}
		for _, v := range node {
			s.Runtimes.Node.Upsert(v, status)
		}
		if ok {
			s.Runtimes.PHP.Default = HighestVersion(php)
			s.Runtimes.Node.Default = HighestVersion(node)
			s.Runtimes.Addons = addons
			s.ProvisionStatus = constants.ProvisionStatusReady
			now := time.Now().UTC()
			s.ProvisionedAt = &now
			if len(secrets) > 0 {
				s.ProvisionSecrets = secrets
			}
		} else {
			s.ProvisionStatus = constants.ProvisionStatusFailed
		}
		return s.SaveProvision(ctx)

	case constants.JobKindRuntimePHPInstall, constants.JobKindRuntimeNodeInstall:
		runtime := constants.RuntimePHP
		if job.Kind == constants.JobKindRuntimeNodeInstall {
			runtime = constants.RuntimeNode
		}
		version, _ := job.Payload["version"].(string)
		family := s.Runtimes.Family(runtime)
		if family == nil || version == "" {
			return nil
		}
		if ok {
			family.Upsert(version, constants.RuntimeStatusInstalled)
			if setDefault, _ := job.Payload["set_default"].(bool); setDefault || family.Default == "" {
				family.Default = version
			}
		} else {
			family.Upsert(version, constants.RuntimeStatusFailed)
		}
		return s.SaveProvision(ctx)

	case constants.JobKindRuntimePHPDefault, constants.JobKindRuntimeNodeDefault:
		if !ok {
			return nil
		}
		runtime := constants.RuntimePHP
		if job.Kind == constants.JobKindRuntimeNodeDefault {
			runtime = constants.RuntimeNode
		}
		version, _ := job.Payload["version"].(string)
		family := s.Runtimes.Family(runtime)
		if family == nil || version == "" {
			return nil
		}
		if v := family.Find(version); v == nil || v.Status != constants.RuntimeStatusInstalled {
			return nil
		}
		family.Default = version
		return s.SaveProvision(ctx)

	case constants.JobKindRuntimePHPRemove, constants.JobKindRuntimeNodeRemove:
		if !ok {
			return nil
		}
		runtime := constants.RuntimePHP
		if job.Kind == constants.JobKindRuntimeNodeRemove {
			runtime = constants.RuntimeNode
		}
		version, _ := job.Payload["version"].(string)
		family := s.Runtimes.Family(runtime)
		if family == nil || version == "" {
			return nil
		}
		family.Remove(version)
		return s.SaveProvision(ctx)
	}

	return nil
}
