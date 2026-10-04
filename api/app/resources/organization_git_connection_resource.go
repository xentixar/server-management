package resources

import "api/app/models"

func OrganizationGitConnectionResource(c models.OrganizationGitConnection) map[string]any {
	var expires any
	if c.TokenExpiresAt != nil {
		expires = c.TokenExpiresAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	return map[string]any{
		"id":                  c.PublicID(),
		"provider":            c.Provider,
		"external_account_id": c.ExternalAccountID,
		"account_login":       c.AccountLogin,
		"account_name":        c.AccountName,
		"avatar_url":          c.AvatarURL,
		"scopes":              c.Scopes,
		"status":              c.Status,
		"last_error":          c.LastError,
		"token_expires_at":    expires,
		"created_at":          c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		"updated_at":          c.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func OrganizationGitConnectionCollection(rows []models.OrganizationGitConnection) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, OrganizationGitConnectionResource(row))
	}
	return out
}
