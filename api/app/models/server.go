package models

import (
	"context"
	"errors"

	"api/app/constants"
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
	if s.SSHPort == 0 {
		s.SSHPort = 22
	}
	if s.Labels == nil {
		s.Labels = []string{}
	}

	return gorm.G[Server](config.GetDatabase()).Create(ctx, s)
}

func (s *Server) Update(ctx context.Context) error {
	if s.AuthenticationMethod != "" && !constants.IsValidAuthMethod(s.AuthenticationMethod) {
		return errors.New("invalid authentication method")
	}
	if s.AgentStatus != "" && !constants.IsValidAgentStatus(s.AgentStatus) {
		return errors.New("invalid agent status")
	}

	_, err := gorm.G[Server](config.GetDatabase()).Where("id = ?", s.ID).Updates(ctx, *s)
	return err
}

func (s *Server) Delete(ctx context.Context) error {
	_, err := gorm.G[Server](config.GetDatabase()).Where("id = ?", s.ID).Delete(ctx)
	return err
}
