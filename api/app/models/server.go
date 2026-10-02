package models

import (
	"context"
	"errors"
	"time"

	"api/app/constants"
	"api/app/helpers"
	"api/config"

	"gorm.io/gorm"
)

type Server struct {
	gorm.Model
	OrganizationID       uint   `gorm:"not null;uniqueIndex:idx_org_hostname"`
	Hostname             string `gorm:"not null;uniqueIndex:idx_org_hostname"`
	IPAddress            string `gorm:"not null"`
	SSHPort              int    `gorm:"not null;default:22"`
	SSHUser              string `gorm:"not null"`
	AuthenticationMethod string `gorm:"type:varchar(20);not null"`
	PrivateKey           string `gorm:"type:text"`
	Passphrase           string `gorm:"type:text"`
	Password             string `gorm:"type:text"`
	Region               string
	Labels               []string `gorm:"serializer:json"`
	Notes                string   `gorm:"type:text"`
	AgentStatus          string   `gorm:"type:varchar(20);not null;default:pending"`
	AgentToken           string   `gorm:"uniqueIndex"`
	AgentVersion         string
	LastSeenAt           *time.Time
	ProvisionStatus      string `gorm:"type:varchar(20);not null;default:none"`
	ProvisionedAt        *time.Time
	Runtimes             ServerRuntimes `gorm:"serializer:json"`
	ProvisionSecrets     map[string]any `gorm:"serializer:json"`

	Organization Organization `gorm:"foreignKey:OrganizationID"`
}

func (Server) FindByID(ctx context.Context, id uint, organizationID uint, withRelations ...string) (Server, error) {
	query := gorm.G[Server](config.GetDatabase()).
		Where("id = ? AND organization_id = ?", id, organizationID)

	for _, relation := range withRelations {
		query = query.Preload(relation, nil)
	}

	return query.First(ctx)
}

func (Server) FindByOrganization(ctx context.Context, organizationID uint) ([]Server, error) {
	return gorm.G[Server](config.GetDatabase()).
		Where("organization_id = ?", organizationID).
		Find(ctx)
}

func (Server) FindByAgentToken(ctx context.Context, token string) (Server, error) {
	return gorm.G[Server](config.GetDatabase()).
		Where("agent_token = ?", token).
		First(ctx)
}

func (s *Server) Create(ctx context.Context) error {
	if !constants.IsValidAuthMethod(s.AuthenticationMethod) {
		return errors.New("invalid authentication method")
	}
	if s.AgentStatus == "" {
		s.AgentStatus = constants.AgentStatusPending
	}
	if !constants.IsValidAgentStatus(s.AgentStatus) {
		return errors.New("invalid agent status")
	}
	if s.ProvisionStatus == "" {
		s.ProvisionStatus = constants.ProvisionStatusNone
	}
	if !constants.IsValidProvisionStatus(s.ProvisionStatus) {
		return errors.New("invalid provision status")
	}
	if s.SSHPort == 0 {
		s.SSHPort = 22
	}
	if s.Labels == nil {
		s.Labels = []string{}
	}
	s.Runtimes.Normalize()
	if s.AgentToken == "" {
		token, err := helpers.GenerateToken()
		if err != nil {
			return err
		}
		s.AgentToken = token
	}

	return gorm.G[Server](config.GetDatabase()).Create(ctx, s)
}

func (s *Server) RecordHeartbeat(ctx context.Context, agentVersion string) error {
	now := time.Now().UTC()
	s.AgentStatus = constants.AgentStatusOnline
	s.AgentVersion = agentVersion
	s.LastSeenAt = &now

	return config.GetDatabase().WithContext(ctx).
		Model(s).
		Updates(map[string]any{
			"agent_status":  s.AgentStatus,
			"agent_version": s.AgentVersion,
			"last_seen_at":  s.LastSeenAt,
		}).Error
}

func (s *Server) Update(ctx context.Context) error {
	if s.AuthenticationMethod != "" && !constants.IsValidAuthMethod(s.AuthenticationMethod) {
		return errors.New("invalid authentication method")
	}
	if s.AgentStatus != "" && !constants.IsValidAgentStatus(s.AgentStatus) {
		return errors.New("invalid agent status")
	}
	if s.ProvisionStatus != "" && !constants.IsValidProvisionStatus(s.ProvisionStatus) {
		return errors.New("invalid provision status")
	}
	s.Runtimes.Normalize()

	_, err := gorm.G[Server](config.GetDatabase()).Where("id = ?", s.ID).Updates(ctx, *s)
	return err
}

func (s *Server) SaveProvision(ctx context.Context) error {
	s.Runtimes.Normalize()
	// Select+struct Updates so serializer:json runs (map[string]any skips it).
	return config.GetDatabase().WithContext(ctx).
		Model(s).
		Select("ProvisionStatus", "ProvisionedAt", "Runtimes", "ProvisionSecrets").
		Updates(s).Error
}

func (s *Server) ClearProvisionSecrets(ctx context.Context) error {
	s.ProvisionSecrets = nil
	return config.GetDatabase().WithContext(ctx).
		Model(s).
		Update("provision_secrets", nil).Error
}

func (s *Server) HasProvisionSecrets() bool {
	return len(s.ProvisionSecrets) > 0
}

func (s *Server) Delete(ctx context.Context) error {
	_, err := gorm.G[Server](config.GetDatabase()).Where("id = ?", s.ID).Delete(ctx)
	return err
}
