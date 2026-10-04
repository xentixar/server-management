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

type OrganizationGitConnection struct {
	gorm.Model
	UID               string `gorm:"type:varchar(36);index"`
	OrganizationID    uint   `gorm:"not null;uniqueIndex:idx_org_git_provider_account"`
	Provider          string `gorm:"type:varchar(32);not null;uniqueIndex:idx_org_git_provider_account"`
	ExternalAccountID string `gorm:"type:varchar(128);not null;uniqueIndex:idx_org_git_provider_account"`
	AccountLogin      string `gorm:"type:varchar(255);not null"`
	AccountName       string `gorm:"type:varchar(255)"`
	AvatarURL         string `gorm:"type:text"`
	AccessTokenEnc    string `gorm:"type:text;not null"`
	RefreshTokenEnc   string `gorm:"type:text"`
	TokenExpiresAt    *time.Time
	Scopes            string `gorm:"type:varchar(512)"`
	Status            string `gorm:"type:varchar(20);not null;default:active"`
	LastError         string `gorm:"type:text"`

	Organization Organization `gorm:"foreignKey:OrganizationID"`
}

func (c OrganizationGitConnection) PublicID() string {
	return c.UID
}

func (c *OrganizationGitConnection) Normalize() {
	c.Provider = strings.TrimSpace(strings.ToLower(c.Provider))
	c.ExternalAccountID = strings.TrimSpace(c.ExternalAccountID)
	c.AccountLogin = strings.TrimSpace(c.AccountLogin)
	c.AccountName = strings.TrimSpace(c.AccountName)
	c.AvatarURL = strings.TrimSpace(c.AvatarURL)
	c.Scopes = strings.TrimSpace(c.Scopes)
	c.Status = strings.TrimSpace(c.Status)
	c.LastError = strings.TrimSpace(c.LastError)
	if c.Status == "" {
		c.Status = constants.GitConnectionStatusActive
	}
}

func (c *OrganizationGitConnection) Validate() error {
	c.Normalize()
	if !constants.IsValidGitProvider(c.Provider) {
		return errors.New("invalid provider")
	}
	if c.OrganizationID == 0 {
		return errors.New("organization required")
	}
	if c.ExternalAccountID == "" {
		return errors.New("external account required")
	}
	if c.AccountLogin == "" {
		return errors.New("account login required")
	}
	if c.AccessTokenEnc == "" {
		return errors.New("access token required")
	}
	if !constants.IsValidGitConnectionStatus(c.Status) {
		return errors.New("invalid status")
	}
	return nil
}

func (c *OrganizationGitConnection) Create(ctx context.Context) error {
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
	return gorm.G[OrganizationGitConnection](config.GetDatabase()).Create(ctx, c)
}

func (c *OrganizationGitConnection) Save(ctx context.Context) error {
	if err := c.Validate(); err != nil {
		return err
	}
	return config.GetDatabase().WithContext(ctx).Save(c).Error
}

func (c *OrganizationGitConnection) Delete(ctx context.Context) error {
	_, err := gorm.G[OrganizationGitConnection](config.GetDatabase()).Where("id = ?", c.ID).Delete(ctx)
	return err
}

func (OrganizationGitConnection) ListByOrganization(ctx context.Context, organizationID uint) ([]OrganizationGitConnection, error) {
	return gorm.G[OrganizationGitConnection](config.GetDatabase()).
		Where("organization_id = ?", organizationID).
		Order("provider asc, account_login asc").
		Find(ctx)
}

func (OrganizationGitConnection) FindByPublicID(ctx context.Context, publicID string, organizationID uint) (OrganizationGitConnection, error) {
	if publicID == "" {
		return OrganizationGitConnection{}, gorm.ErrRecordNotFound
	}
	return gorm.G[OrganizationGitConnection](config.GetDatabase()).
		Where("uid = ? AND organization_id = ?", publicID, organizationID).
		First(ctx)
}

func (OrganizationGitConnection) FindByProviderAccount(ctx context.Context, organizationID uint, provider, externalAccountID string) (OrganizationGitConnection, error) {
	return gorm.G[OrganizationGitConnection](config.GetDatabase()).
		Where(
			"organization_id = ? AND provider = ? AND external_account_id = ?",
			organizationID, provider, externalAccountID,
		).
		First(ctx)
}

// GitOAuthState is a short-lived CSRF token for the OAuth redirect dance.
type GitOAuthState struct {
	gorm.Model
	State          string    `gorm:"type:varchar(64);uniqueIndex;not null"`
	OrganizationID uint      `gorm:"not null;index"`
	UserID         uint      `gorm:"not null"`
	Provider       string    `gorm:"type:varchar(32);not null"`
	ExpiresAt      time.Time `gorm:"not null;index"`

	Organization Organization `gorm:"foreignKey:OrganizationID"`
}

func (s *GitOAuthState) Create(ctx context.Context) error {
	s.Provider = strings.TrimSpace(strings.ToLower(s.Provider))
	s.State = strings.TrimSpace(s.State)
	if s.State == "" || s.OrganizationID == 0 || s.UserID == 0 || !constants.IsValidGitProvider(s.Provider) {
		return errors.New("invalid oauth state")
	}
	if s.ExpiresAt.IsZero() {
		s.ExpiresAt = time.Now().UTC().Add(10 * time.Minute)
	}
	return gorm.G[GitOAuthState](config.GetDatabase()).Create(ctx, s)
}

func (GitOAuthState) Consume(ctx context.Context, state, provider string) (GitOAuthState, error) {
	state = strings.TrimSpace(state)
	provider = strings.TrimSpace(strings.ToLower(provider))
	if state == "" {
		return GitOAuthState{}, gorm.ErrRecordNotFound
	}
	row, err := gorm.G[GitOAuthState](config.GetDatabase()).
		Where("state = ? AND provider = ?", state, provider).
		First(ctx)
	if err != nil {
		return GitOAuthState{}, err
	}
	_, _ = gorm.G[GitOAuthState](config.GetDatabase()).Where("id = ?", row.ID).Delete(ctx)
	if time.Now().UTC().After(row.ExpiresAt) {
		return GitOAuthState{}, errors.New("oauth state expired")
	}
	return row, nil
}

func (GitOAuthState) PurgeExpired(ctx context.Context) error {
	_, err := gorm.G[GitOAuthState](config.GetDatabase()).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(ctx)
	return err
}
