package models

import (
	"context"
	"errors"
	"strings"

	"api/app/constants"
	"api/app/helpers"
	"api/config"

	"gorm.io/gorm"
)

type ServerService struct {
	gorm.Model
	UID         string `gorm:"type:varchar(36);index"`
	ServerID    uint   `gorm:"not null;uniqueIndex:idx_server_service_name"`
	Name        string `gorm:"type:varchar(64);not null;uniqueIndex:idx_server_service_name"`
	Command     string `gorm:"type:text;not null"`
	Directory   string `gorm:"type:text"`
	User        string `gorm:"type:varchar(64);not null;default:sm"`
	Processes   int    `gorm:"not null;default:1"`
	AutoRestart bool   `gorm:"not null;default:true"`
	Autostart   bool   `gorm:"not null;default:true"`
	Enabled     bool   `gorm:"not null;default:true"`
	Status      string `gorm:"type:varchar(20);not null;default:pending"`
	StdoutTail  string `gorm:"type:text"`
	StderrTail  string `gorm:"type:text"`
	Message     string `gorm:"type:text"`

	Server Server `gorm:"foreignKey:ServerID"`
}

func (s ServerService) PublicID() string {
	if s.UID != "" {
		return s.UID
	}
	return ""
}

func (s *ServerService) Normalize() {
	s.Name = strings.TrimSpace(s.Name)
	s.Command = strings.TrimSpace(s.Command)
	s.Directory = strings.TrimSpace(s.Directory)
	s.User = strings.TrimSpace(s.User)
	if s.User == "" {
		s.User = constants.DefaultProcessUser
	}
	if s.Processes < 1 {
		s.Processes = 1
	}
	if s.Processes > 64 {
		s.Processes = 64
	}
	if s.Status == "" {
		s.Status = constants.ServiceStatusPending
	}
}

func (s *ServerService) Validate() error {
	s.Normalize()
	if !helpers.IsValidServiceName(s.Name) {
		return errors.New("invalid service name")
	}
	if !helpers.IsValidProcessCommand(s.Command) {
		return errors.New("invalid command")
	}
	if !constants.IsValidServiceStatus(s.Status) {
		return errors.New("invalid status")
	}
	return nil
}

func (s *ServerService) Create(ctx context.Context) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if s.UID == "" {
		uid, err := helpers.GenerateUID()
		if err != nil {
			return err
		}
		s.UID = uid
	}
	return gorm.G[ServerService](config.GetDatabase()).Create(ctx, s)
}

func (s *ServerService) Save(ctx context.Context) error {
	if err := s.Validate(); err != nil {
		return err
	}
	return config.GetDatabase().WithContext(ctx).Save(s).Error
}

func (s *ServerService) Delete(ctx context.Context) error {
	_, err := gorm.G[ServerService](config.GetDatabase()).Where("id = ?", s.ID).Delete(ctx)
	return err
}

func (ServerService) ListByServer(ctx context.Context, serverID uint) ([]ServerService, error) {
	return gorm.G[ServerService](config.GetDatabase()).
		Where("server_id = ?", serverID).
		Order("name asc").
		Find(ctx)
}

func (ServerService) FindByPublicID(ctx context.Context, publicID string, serverID uint) (ServerService, error) {
	if publicID == "" {
		return ServerService{}, gorm.ErrRecordNotFound
	}
	return gorm.G[ServerService](config.GetDatabase()).
		Where("uid = ? AND server_id = ?", publicID, serverID).
		First(ctx)
}

func (ServerService) FindByName(ctx context.Context, serverID uint, name string) (ServerService, error) {
	return gorm.G[ServerService](config.GetDatabase()).
		Where("server_id = ? AND name = ?", serverID, name).
		First(ctx)
}

func (ServerService) DeleteForServer(ctx context.Context, serverID uint) error {
	_, err := gorm.G[ServerService](config.GetDatabase()).
		Where("server_id = ?", serverID).
		Delete(ctx)
	return err
}

func (s ServerService) JobPayload() map[string]any {
	return map[string]any{
		"uid":          s.PublicID(),
		"name":         s.Name,
		"command":      s.Command,
		"directory":    s.Directory,
		"user":         s.User,
		"processes":    s.Processes,
		"auto_restart": s.AutoRestart,
		"autostart":    s.Autostart,
		"enabled":      s.Enabled,
	}
}
