package resources

import (
	"api/app/models"
)

func AgentJobResource(job models.AgentJob) map[string]any {
	payload := job.Payload
	if payload == nil {
		payload = map[string]any{}
	}

	var claimedAt any
	if job.ClaimedAt != nil {
		claimedAt = job.ClaimedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	var finishedAt any
	if job.FinishedAt != nil {
		finishedAt = job.FinishedAt.UTC().Format("2006-01-02T15:04:05Z")
	}

	return map[string]any{
		"id":          job.PublicID(),
		"kind":        job.Kind,
		"payload":     payload,
		"status":      job.Status,
		"message":     job.Message,
		"exit_code":   job.ExitCode,
		"claimed_at":  claimedAt,
		"finished_at": finishedAt,
		"created_at":  job.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
