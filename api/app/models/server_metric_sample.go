package models

import (
	"context"
	"time"

	"api/config"
)

const MetricRetention = 7 * 24 * time.Hour

// ServerMetricSample is one host metrics snapshot from an agent heartbeat.
type ServerMetricSample struct {
	ID             uint      `gorm:"primaryKey"`
	ServerID       uint      `gorm:"not null;index:idx_server_metric_recorded,priority:1"`
	RecordedAt     time.Time `gorm:"not null;index:idx_server_metric_recorded,priority:2"`
	CPUPercent     *float64
	MemUsedBytes   *uint64
	MemTotalBytes  *uint64
	DiskUsedBytes  *uint64
	DiskTotalBytes *uint64
	Load1          *float64
	Load5          *float64
	Load15         *float64
}

func (ServerMetricSample) TableName() string {
	return "server_metric_samples"
}

// HostMetricsPayload is the optional metrics object on agent heartbeat.
type HostMetricsPayload struct {
	CPUPercent     *float64 `json:"cpu_percent"`
	MemUsedBytes   *uint64  `json:"mem_used_bytes"`
	MemTotalBytes  *uint64  `json:"mem_total_bytes"`
	DiskUsedBytes  *uint64  `json:"disk_used_bytes"`
	DiskTotalBytes *uint64  `json:"disk_total_bytes"`
	Load1          *float64 `json:"load1"`
	Load5          *float64 `json:"load5"`
	Load15         *float64 `json:"load15"`
}

func (p *HostMetricsPayload) Empty() bool {
	if p == nil {
		return true
	}
	return p.CPUPercent == nil && p.MemTotalBytes == nil && p.DiskTotalBytes == nil &&
		p.Load1 == nil && p.Load5 == nil && p.Load15 == nil
}

func (ServerMetricSample) InsertFromHeartbeat(ctx context.Context, serverID uint, at time.Time, p *HostMetricsPayload) error {
	if p.Empty() {
		return nil
	}
	sample := ServerMetricSample{
		ServerID:       serverID,
		RecordedAt:     at.UTC(),
		CPUPercent:     p.CPUPercent,
		MemUsedBytes:   p.MemUsedBytes,
		MemTotalBytes:  p.MemTotalBytes,
		DiskUsedBytes:  p.DiskUsedBytes,
		DiskTotalBytes: p.DiskTotalBytes,
		Load1:          p.Load1,
		Load5:          p.Load5,
		Load15:         p.Load15,
	}
	db := config.GetDatabase().WithContext(ctx)
	if err := db.Create(&sample).Error; err != nil {
		return err
	}
	cutoff := at.UTC().Add(-MetricRetention)
	return db.Where("server_id = ? AND recorded_at < ?", serverID, cutoff).
		Delete(&ServerMetricSample{}).Error
}

func (ServerMetricSample) DeleteForServer(ctx context.Context, serverID uint) error {
	return config.GetDatabase().WithContext(ctx).
		Where("server_id = ?", serverID).
		Delete(&ServerMetricSample{}).Error
}

func (ServerMetricSample) Latest(ctx context.Context, serverID uint) (*ServerMetricSample, error) {
	var sample ServerMetricSample
	err := config.GetDatabase().WithContext(ctx).
		Where("server_id = ?", serverID).
		Order("recorded_at DESC").
		Limit(1).
		First(&sample).Error
	if err != nil {
		return nil, err
	}
	return &sample, nil
}

// MetricSeriesPoint is a (possibly bucketed) point for charts.
type MetricSeriesPoint struct {
	T              time.Time `json:"t"`
	CPUPercent     *float64  `json:"cpu_percent"`
	MemUsedBytes   *float64  `json:"mem_used_bytes"`
	MemTotalBytes  *float64  `json:"mem_total_bytes"`
	DiskUsedBytes  *float64  `json:"disk_used_bytes"`
	DiskTotalBytes *float64  `json:"disk_total_bytes"`
	Load1          *float64  `json:"load1"`
	Load5          *float64  `json:"load5"`
	Load15         *float64  `json:"load15"`
}

// Series returns samples from [from, to], optionally bucketed with Postgres date_trunc.
// bucket is "" for raw rows, or "minute" / "hour".
func (ServerMetricSample) Series(ctx context.Context, serverID uint, from, to time.Time, bucket string) ([]MetricSeriesPoint, error) {
	db := config.GetDatabase().WithContext(ctx)
	from = from.UTC()
	to = to.UTC()

	if bucket == "" {
		var rows []ServerMetricSample
		err := db.Where("server_id = ? AND recorded_at >= ? AND recorded_at <= ?", serverID, from, to).
			Order("recorded_at ASC").
			Find(&rows).Error
		if err != nil {
			return nil, err
		}
		out := make([]MetricSeriesPoint, 0, len(rows))
		for _, r := range rows {
			out = append(out, MetricSeriesPoint{
				T:              r.RecordedAt.UTC(),
				CPUPercent:     r.CPUPercent,
				MemUsedBytes:   uint64PtrToFloat(r.MemUsedBytes),
				MemTotalBytes:  uint64PtrToFloat(r.MemTotalBytes),
				DiskUsedBytes:  uint64PtrToFloat(r.DiskUsedBytes),
				DiskTotalBytes: uint64PtrToFloat(r.DiskTotalBytes),
				Load1:          r.Load1,
				Load5:          r.Load5,
				Load15:         r.Load15,
			})
		}
		return out, nil
	}

	// AVG over nullable columns; cast bigint to float for avg.
	type aggRow struct {
		Bucket         time.Time `gorm:"column:bucket"`
		CPUPercent     *float64  `gorm:"column:cpu_percent"`
		MemUsedBytes   *float64  `gorm:"column:mem_used_bytes"`
		MemTotalBytes  *float64  `gorm:"column:mem_total_bytes"`
		DiskUsedBytes  *float64  `gorm:"column:disk_used_bytes"`
		DiskTotalBytes *float64  `gorm:"column:disk_total_bytes"`
		Load1          *float64  `gorm:"column:load1"`
		Load5          *float64  `gorm:"column:load5"`
		Load15         *float64  `gorm:"column:load15"`
	}
	var rows []aggRow
	sql := `
SELECT date_trunc(?, recorded_at) AS bucket,
       AVG(cpu_percent) AS cpu_percent,
       AVG(mem_used_bytes::double precision) AS mem_used_bytes,
       AVG(mem_total_bytes::double precision) AS mem_total_bytes,
       AVG(disk_used_bytes::double precision) AS disk_used_bytes,
       AVG(disk_total_bytes::double precision) AS disk_total_bytes,
       AVG(load1) AS load1,
       AVG(load5) AS load5,
       AVG(load15) AS load15
FROM server_metric_samples
WHERE server_id = ? AND recorded_at >= ? AND recorded_at <= ?
GROUP BY 1
ORDER BY 1 ASC`
	if err := db.Raw(sql, bucket, serverID, from, to).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]MetricSeriesPoint, 0, len(rows))
	for _, r := range rows {
		out = append(out, MetricSeriesPoint{
			T:              r.Bucket.UTC(),
			CPUPercent:     r.CPUPercent,
			MemUsedBytes:   r.MemUsedBytes,
			MemTotalBytes:  r.MemTotalBytes,
			DiskUsedBytes:  r.DiskUsedBytes,
			DiskTotalBytes: r.DiskTotalBytes,
			Load1:          r.Load1,
			Load5:          r.Load5,
			Load15:         r.Load15,
		})
	}
	return out, nil
}

func uint64PtrToFloat(v *uint64) *float64 {
	if v == nil {
		return nil
	}
	f := float64(*v)
	return &f
}
