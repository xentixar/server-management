package resources

import (
	"api/app/models"
)

func ServerMetricSampleResource(s *models.ServerMetricSample) map[string]any {
	if s == nil {
		return nil
	}
	return map[string]any{
		"t":                s.RecordedAt.UTC().Format("2006-01-02T15:04:05Z"),
		"cpu_percent":      s.CPUPercent,
		"mem_used_bytes":   s.MemUsedBytes,
		"mem_total_bytes":  s.MemTotalBytes,
		"disk_used_bytes":  s.DiskUsedBytes,
		"disk_total_bytes": s.DiskTotalBytes,
		"load1":            s.Load1,
		"load5":            s.Load5,
		"load15":           s.Load15,
	}
}

func ServerMetricSeriesResource(points []models.MetricSeriesPoint) []map[string]any {
	out := make([]map[string]any, 0, len(points))
	for _, p := range points {
		out = append(out, map[string]any{
			"t":                p.T.UTC().Format("2006-01-02T15:04:05Z"),
			"cpu_percent":      p.CPUPercent,
			"mem_used_bytes":   p.MemUsedBytes,
			"mem_total_bytes":  p.MemTotalBytes,
			"disk_used_bytes":  p.DiskUsedBytes,
			"disk_total_bytes": p.DiskTotalBytes,
			"load1":            p.Load1,
			"load5":            p.Load5,
			"load15":           p.Load15,
		})
	}
	return out
}
