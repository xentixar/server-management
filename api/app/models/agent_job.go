package models

import (
	"context"
	"errors"
	"strconv"
	"time"

	"api/app/constants"
	"api/app/helpers"
	"api/config"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type AgentJob struct {
	gorm.Model
	UID        string         `gorm:"type:varchar(36);index"` // opaque public id; unique in practice via GenerateUID
	ServerID   uint           `gorm:"not null;index"`
	Kind       string         `gorm:"type:varchar(64);not null"`
	Payload    map[string]any `gorm:"serializer:json"`
	Status     string         `gorm:"type:varchar(20);not null;index"`
	Message    string         `gorm:"type:text"`
	ExitCode   int
	ClaimedAt  *time.Time
	FinishedAt *time.Time

	Server Server `gorm:"foreignKey:ServerID"`
}

func (j AgentJob) PublicID() string {
	if j.UID != "" {
		return j.UID
	}
	return strconv.FormatUint(uint64(j.ID), 10)
}

func (AgentJob) Enqueue(ctx context.Context, serverID uint, kind string, payload map[string]any) (AgentJob, error) {
	if kind == "" {
		return AgentJob{}, errors.New("job kind is required")
	}
	if payload == nil {
		payload = map[string]any{}
	}
	uid, err := helpers.GenerateUID()
	if err != nil {
		return AgentJob{}, err
	}

	job := AgentJob{
		UID:      uid,
		ServerID: serverID,
		Kind:     kind,
		Payload:  payload,
		Status:   constants.JobStatusPending,
	}
	if err := gorm.G[AgentJob](config.GetDatabase()).Create(ctx, &job); err != nil {
		return AgentJob{}, err
	}
	return job, nil
}

// ClaimNext marks the oldest pending job for the server as running.
func (AgentJob) ClaimNext(ctx context.Context, serverID uint) (*AgentJob, error) {
	db := config.GetDatabase()

	var job AgentJob
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("server_id = ? AND status = ?", serverID, constants.JobStatusPending).
			Order("id asc").
			First(&job).Error
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		job.Status = constants.JobStatusRunning
		job.ClaimedAt = &now
		return tx.Model(&job).Select("Status", "ClaimedAt").Updates(job).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (AgentJob) FindForServer(ctx context.Context, id uint, serverID uint) (AgentJob, error) {
	return gorm.G[AgentJob](config.GetDatabase()).
		Where("id = ? AND server_id = ?", id, serverID).
		First(ctx)
}

func (AgentJob) FindByPublicID(ctx context.Context, publicID string, serverID uint) (AgentJob, error) {
	if publicID == "" {
		return AgentJob{}, gorm.ErrRecordNotFound
	}
	db := gorm.G[AgentJob](config.GetDatabase())
	job, err := db.Where("uid = ? AND server_id = ?", publicID, serverID).First(ctx)
	if err == nil {
		return job, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return AgentJob{}, err
	}
	id, perr := strconv.ParseUint(publicID, 10, 64)
	if perr != nil || id == 0 {
		return AgentJob{}, gorm.ErrRecordNotFound
	}
	return (AgentJob{}).FindForServer(ctx, uint(id), serverID)
}

func (AgentJob) HasOpen(ctx context.Context, serverID uint, kind string) (bool, error) {
	var count int64
	err := config.GetDatabase().WithContext(ctx).
		Model(&AgentJob{}).
		Where("server_id = ? AND kind = ? AND status IN ?", serverID, kind, []string{
			constants.JobStatusPending,
			constants.JobStatusRunning,
		}).
		Count(&count).Error
	return count > 0, err
}

func (AgentJob) LatestByKind(ctx context.Context, serverID uint, kind string) (*AgentJob, error) {
	var job AgentJob
	q := config.GetDatabase().WithContext(ctx).Where("server_id = ?", serverID)
	if kind != "" {
		q = q.Where("kind = ?", kind)
	}
	err := q.Order("id desc").First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (AgentJob) DeleteForServer(ctx context.Context, serverID uint) error {
	_, err := gorm.G[AgentJob](config.GetDatabase()).
		Where("server_id = ?", serverID).
		Delete(ctx)
	return err
}

// BackfillUIDs assigns opaque uids to legacy rows (needed before unique index is clean).
func (AgentJob) BackfillUIDs() error {
	db := config.GetDatabase()
	var jobs []AgentJob
	if err := db.Where("uid = ? OR uid IS NULL", "").Find(&jobs).Error; err != nil {
		return err
	}
	for i := range jobs {
		uid, err := helpers.GenerateUID()
		if err != nil {
			return err
		}
		if err := db.Model(&jobs[i]).Update("uid", uid).Error; err != nil {
			return err
		}
	}
	return nil
}

func (j *AgentJob) Complete(ctx context.Context, ok bool, message string, exitCode int) error {
	if j.Status != constants.JobStatusRunning && j.Status != constants.JobStatusPending {
		return errors.New("job is not open")
	}

	now := time.Now().UTC()
	if ok {
		j.Status = constants.JobStatusSucceeded
	} else {
		j.Status = constants.JobStatusFailed
	}
	j.Message = message
	j.ExitCode = exitCode
	j.FinishedAt = &now

	return config.GetDatabase().WithContext(ctx).
		Model(j).
		Updates(map[string]any{
			"status":      j.Status,
			"message":     j.Message,
			"exit_code":   j.ExitCode,
			"finished_at": j.FinishedAt,
		}).Error
}
