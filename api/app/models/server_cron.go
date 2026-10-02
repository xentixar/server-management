package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"api/app/constants"
	"api/app/helpers"
	"api/config"

	"gorm.io/gorm"
)

type ServerCron struct {
	gorm.Model
	UID      string `gorm:"type:varchar(36);index"`
	ServerID uint   `gorm:"not null;index"`
	Schedule string `gorm:"type:varchar(128);not null"`
	Command  string `gorm:"type:text;not null"`
	User     string `gorm:"type:varchar(64);not null;default:sm"`
	Timezone string `gorm:"type:varchar(64);not null;default:UTC"`
	Enabled  bool   `gorm:"not null;default:true"`
	Status   string `gorm:"type:varchar(20);not null;default:pending"`
	Message  string `gorm:"type:text"`
	LogTail  string `gorm:"type:text"`

	Server Server `gorm:"foreignKey:ServerID"`
}

func (c ServerCron) PublicID() string {
	if c.UID != "" {
		return c.UID
	}
	return ""
}

func (c *ServerCron) Normalize() {
	c.Schedule = strings.TrimSpace(c.Schedule)
	c.Command = strings.TrimSpace(c.Command)
	c.User = strings.TrimSpace(c.User)
	c.Timezone = strings.TrimSpace(c.Timezone)
	if c.User == "" {
		c.User = constants.DefaultProcessUser
	}
	if c.Timezone == "" {
		c.Timezone = constants.DefaultCronTZ
	}
	if c.Status == "" {
		c.Status = constants.CronStatusPending
	}
}

func (c *ServerCron) Validate() error {
	c.Normalize()
	if !helpers.IsValidCronSchedule(c.Schedule) {
		return errors.New("invalid cron schedule")
	}
	if !helpers.IsValidProcessCommand(c.Command) {
		return errors.New("invalid command")
	}
	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return errors.New("invalid timezone")
	}
	if !constants.IsValidCronStatus(c.Status) {
		return errors.New("invalid status")
	}
	return nil
}

func (c *ServerCron) Create(ctx context.Context) error {
	if err := c.Validate(); err != nil {
		return err
	}
	if c.UID == "" {
		uid, err := helpers.GenerateUID()
		if err != nil {
			return err
		}
		c.UID = uid
	}
	return gorm.G[ServerCron](config.GetDatabase()).Create(ctx, c)
}

func (c *ServerCron) Save(ctx context.Context) error {
	if err := c.Validate(); err != nil {
		return err
	}
	return config.GetDatabase().WithContext(ctx).Save(c).Error
}

func (c *ServerCron) Delete(ctx context.Context) error {
	_, err := gorm.G[ServerCron](config.GetDatabase()).Where("id = ?", c.ID).Delete(ctx)
	return err
}

func (ServerCron) ListByServer(ctx context.Context, serverID uint) ([]ServerCron, error) {
	return gorm.G[ServerCron](config.GetDatabase()).
		Where("server_id = ?", serverID).
		Order("id asc").
		Find(ctx)
}

func (ServerCron) FindByPublicID(ctx context.Context, publicID string, serverID uint) (ServerCron, error) {
	if publicID == "" {
		return ServerCron{}, gorm.ErrRecordNotFound
	}
	return gorm.G[ServerCron](config.GetDatabase()).
		Where("uid = ? AND server_id = ?", publicID, serverID).
		First(ctx)
}

func (ServerCron) DeleteForServer(ctx context.Context, serverID uint) error {
	_, err := gorm.G[ServerCron](config.GetDatabase()).
		Where("server_id = ?", serverID).
		Delete(ctx)
	return err
}

func (c ServerCron) SyncItem() map[string]any {
	return map[string]any{
		"uid":      c.PublicID(),
		"schedule": c.Schedule,
		"command":  c.Command,
		"user":     c.User,
		"timezone": c.Timezone,
		"enabled":  c.Enabled,
	}
}

func CronSyncPayload(crons []ServerCron) map[string]any {
	items := make([]map[string]any, 0, len(crons))
	for _, c := range crons {
		if !c.Enabled {
			continue
		}
		items = append(items, c.SyncItem())
	}
	return map[string]any{"entries": items}
}
