package resources

import "api/app/models"

func ServerServiceResource(s models.ServerService) map[string]any {
	return map[string]any{
		"id":           s.PublicID(),
		"name":         s.Name,
		"command":      s.Command,
		"directory":    s.Directory,
		"user":         s.User,
		"processes":    s.Processes,
		"auto_restart": s.AutoRestart,
		"autostart":    s.Autostart,
		"enabled":      s.Enabled,
		"status":       s.Status,
		"stdout_tail":  s.StdoutTail,
		"stderr_tail":  s.StderrTail,
		"message":      s.Message,
		"created_at":   s.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		"updated_at":   s.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func ServerCronResource(c models.ServerCron) map[string]any {
	return map[string]any{
		"id":         c.PublicID(),
		"schedule":   c.Schedule,
		"command":    c.Command,
		"user":       c.User,
		"timezone":   c.Timezone,
		"enabled":    c.Enabled,
		"status":     c.Status,
		"message":    c.Message,
		"created_at": c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		"updated_at": c.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
