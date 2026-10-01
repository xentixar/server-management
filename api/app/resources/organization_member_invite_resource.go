package resources

import "api/app/models"

func OrganizationMemberInviteResource(data models.OrganizationMemberInvite) map[string]any {
	return map[string]any{
		"id":         data.ID,
		"email":      data.Email,
		"role":       data.Role,
		"created_at": data.CreatedAt,
	}
}

func OrganizationMemberInviteCollection(data []models.OrganizationMemberInvite) []map[string]any {
	return Collect(data, OrganizationMemberInviteResource)
}
