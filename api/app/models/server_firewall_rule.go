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

type ServerFirewallRule struct {
	gorm.Model
	UID       string `gorm:"type:varchar(36);index"`
	ServerID  uint   `gorm:"not null;index"`
	Name      string `gorm:"type:varchar(64)"`
	Action    string `gorm:"type:varchar(16);not null;default:allow"`
	Direction string `gorm:"type:varchar(16);not null;default:inbound"`
	Protocol  string `gorm:"type:varchar(16);not null;default:tcp"`
	PortStart int    `gorm:"not null;default:0"`
	PortEnd   int    `gorm:"not null;default:0"`
	Source    string `gorm:"type:varchar(64);not null;default:0.0.0.0/0"`
	Position  int    `gorm:"not null;default:0;index"`
	Enabled   bool   `gorm:"not null;default:true"`
	Status    string `gorm:"type:varchar(20);not null;default:pending"`
	Message   string `gorm:"type:text"`

	Server Server `gorm:"foreignKey:ServerID"`
}

func (r ServerFirewallRule) PublicID() string {
	return r.UID
}

func (r *ServerFirewallRule) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Action = strings.ToLower(strings.TrimSpace(r.Action))
	r.Direction = strings.ToLower(strings.TrimSpace(r.Direction))
	r.Protocol = strings.ToLower(strings.TrimSpace(r.Protocol))
	r.Source = helpers.NormalizeFirewallCIDR(r.Source)
	if r.Action == "" {
		r.Action = constants.FirewallActionAllow
	}
	if r.Direction == "" {
		r.Direction = constants.FirewallDirectionInbound
	}
	if r.Protocol == "" {
		r.Protocol = constants.FirewallProtoTCP
	}
	if r.Status == "" {
		r.Status = constants.FirewallStatusPending
	}
	if r.PortEnd == 0 {
		r.PortEnd = r.PortStart
	}
}

func (r *ServerFirewallRule) Validate() error {
	r.Normalize()
	if !helpers.IsValidFirewallName(r.Name) {
		return errors.New("invalid name")
	}
	if !constants.IsValidFirewallAction(r.Action) {
		return errors.New("invalid action")
	}
	if !constants.IsValidFirewallDirection(r.Direction) {
		return errors.New("invalid direction")
	}
	if !constants.IsValidFirewallProtocol(r.Protocol) {
		return errors.New("invalid protocol")
	}
	if !helpers.IsValidFirewallCIDR(r.Source) {
		return errors.New("invalid source")
	}
	if !constants.IsValidFirewallStatus(r.Status) {
		return errors.New("invalid status")
	}
	switch r.Protocol {
	case constants.FirewallProtoICMP, constants.FirewallProtoAny:
		if r.PortStart != 0 || r.PortEnd != 0 {
			return errors.New("ports are not allowed for icmp/any")
		}
	default:
		if r.PortStart < 0 || r.PortStart > 65535 || r.PortEnd < 0 || r.PortEnd > 65535 {
			return errors.New("port out of range")
		}
		if r.PortStart == 0 && r.PortEnd == 0 {
			return nil
		}
		if r.PortStart == 0 || r.PortEnd == 0 {
			return errors.New("port range incomplete")
		}
		if r.PortEnd < r.PortStart {
			return errors.New("port end before start")
		}
	}
	return nil
}

func (r *ServerFirewallRule) Create(ctx context.Context) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if r.UID == "" {
		uid, err := helpers.GenerateUID()
		if err != nil {
			return err
		}
		r.UID = uid
	}
	return gorm.G[ServerFirewallRule](config.GetDatabase()).Create(ctx, r)
}

func (r *ServerFirewallRule) Save(ctx context.Context) error {
	if err := r.Validate(); err != nil {
		return err
	}
	return config.GetDatabase().WithContext(ctx).Save(r).Error
}

func (r *ServerFirewallRule) Delete(ctx context.Context) error {
	_, err := gorm.G[ServerFirewallRule](config.GetDatabase()).Where("id = ?", r.ID).Delete(ctx)
	return err
}

func (ServerFirewallRule) ListByServer(ctx context.Context, serverID uint) ([]ServerFirewallRule, error) {
	return gorm.G[ServerFirewallRule](config.GetDatabase()).
		Where("server_id = ?", serverID).
		Order("position asc, id asc").
		Find(ctx)
}

func (ServerFirewallRule) FindByPublicID(ctx context.Context, publicID string, serverID uint) (ServerFirewallRule, error) {
	if publicID == "" {
		return ServerFirewallRule{}, gorm.ErrRecordNotFound
	}
	return gorm.G[ServerFirewallRule](config.GetDatabase()).
		Where("uid = ? AND server_id = ?", publicID, serverID).
		First(ctx)
}

func (ServerFirewallRule) NextPosition(ctx context.Context, serverID uint) (int, error) {
	var max *int
	err := config.GetDatabase().WithContext(ctx).
		Model(&ServerFirewallRule{}).
		Select("MAX(position)").
		Where("server_id = ?", serverID).
		Scan(&max).Error
	if err != nil {
		return 0, err
	}
	if max == nil {
		return 1, nil
	}
	return *max + 1, nil
}

func (ServerFirewallRule) DeleteForServer(ctx context.Context, serverID uint) error {
	_, err := gorm.G[ServerFirewallRule](config.GetDatabase()).
		Where("server_id = ?", serverID).
		Delete(ctx)
	return err
}

func (r ServerFirewallRule) SyncItem() map[string]any {
	return map[string]any{
		"uid":       r.PublicID(),
		"name":      r.Name,
		"action":    r.Action,
		"direction": r.Direction,
		"protocol":  r.Protocol,
		"port_start": r.PortStart,
		"port_end":   r.PortEnd,
		"source":    r.Source,
		"position":  r.Position,
		"enabled":   r.Enabled,
	}
}

func FirewallSyncPayload(sshPort int, rules []ServerFirewallRule) map[string]any {
	if sshPort <= 0 {
		sshPort = 22
	}
	items := make([]map[string]any, 0, len(rules))
	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		items = append(items, r.SyncItem())
	}
	return map[string]any{
		"ssh_port": sshPort,
		"rules":    items,
	}
}
