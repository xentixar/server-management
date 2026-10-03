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

type ServerSshKey struct {
	gorm.Model
	UID         string `gorm:"type:varchar(36);index"`
	ServerID    uint   `gorm:"not null;index"`
	Name        string `gorm:"type:varchar(64);not null"`
	PublicKey   string `gorm:"type:text;not null"`
	Fingerprint string `gorm:"type:varchar(128);not null;index"`
	Status      string `gorm:"type:varchar(20);not null;default:pending"`
	Message     string `gorm:"type:text"`

	Server Server `gorm:"foreignKey:ServerID"`
}

func (k ServerSshKey) PublicID() string {
	return k.UID
}

func (k *ServerSshKey) Normalize() {
	k.Name = strings.TrimSpace(k.Name)
	k.PublicKey = strings.TrimSpace(k.PublicKey)
	k.Fingerprint = strings.TrimSpace(k.Fingerprint)
	k.Message = strings.TrimSpace(k.Message)
	if k.Status == "" {
		k.Status = constants.SshKeyStatusPending
	}
}

func (k *ServerSshKey) Validate() error {
	k.Normalize()
	if !helpers.IsValidServiceName(k.Name) {
		return errors.New("invalid name")
	}
	normalized, fingerprint, err := helpers.NormalizePublicKey(k.PublicKey)
	if err != nil {
		return err
	}
	k.PublicKey = normalized
	k.Fingerprint = fingerprint
	if !constants.IsValidSshKeyStatus(k.Status) {
		return errors.New("invalid status")
	}
	return nil
}

func (k *ServerSshKey) Create(ctx context.Context) error {
	if err := k.Validate(); err != nil {
		return err
	}
	if k.UID == "" {
		uid, err := helpers.GenerateUID()
		if err != nil {
			return err
		}
		k.UID = uid
	}
	return gorm.G[ServerSshKey](config.GetDatabase()).Create(ctx, k)
}

func (k *ServerSshKey) Save(ctx context.Context) error {
	if err := k.Validate(); err != nil {
		return err
	}
	return config.GetDatabase().WithContext(ctx).Save(k).Error
}

func (k *ServerSshKey) Delete(ctx context.Context) error {
	_, err := gorm.G[ServerSshKey](config.GetDatabase()).Where("id = ?", k.ID).Delete(ctx)
	return err
}

func (ServerSshKey) ListByServer(ctx context.Context, serverID uint) ([]ServerSshKey, error) {
	return gorm.G[ServerSshKey](config.GetDatabase()).
		Where("server_id = ?", serverID).
		Order("id asc").
		Find(ctx)
}

func (ServerSshKey) FindByPublicID(ctx context.Context, publicID string, serverID uint) (ServerSshKey, error) {
	if publicID == "" {
		return ServerSshKey{}, gorm.ErrRecordNotFound
	}
	return gorm.G[ServerSshKey](config.GetDatabase()).
		Where("uid = ? AND server_id = ?", publicID, serverID).
		First(ctx)
}

func (ServerSshKey) FindByFingerprint(ctx context.Context, serverID uint, fingerprint string, excludeID uint) (ServerSshKey, error) {
	fingerprint = strings.TrimSpace(fingerprint)
	if fingerprint == "" {
		return ServerSshKey{}, gorm.ErrRecordNotFound
	}
	q := gorm.G[ServerSshKey](config.GetDatabase()).
		Where("server_id = ? AND fingerprint = ?", serverID, fingerprint)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	return q.First(ctx)
}

func (ServerSshKey) DeleteForServer(ctx context.Context, serverID uint) error {
	_, err := gorm.G[ServerSshKey](config.GetDatabase()).
		Where("server_id = ?", serverID).
		Delete(ctx)
	return err
}

func (k ServerSshKey) SyncItem() map[string]any {
	return map[string]any{
		"uid":        k.PublicID(),
		"name":       k.Name,
		"public_key": k.PublicKey,
	}
}

func SshKeySyncPayload(keys []ServerSshKey) map[string]any {
	items := make([]map[string]any, 0, len(keys))
	for _, k := range keys {
		items = append(items, k.SyncItem())
	}
	return map[string]any{
		"user": "sm",
		"keys": items,
	}
}
