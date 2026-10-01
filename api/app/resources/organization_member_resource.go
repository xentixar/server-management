package resources

import "api/app/models"

func OrganizationMemberResource(data models.OrganizationMember) map[string]any {
	return map[string]any{
		"id":    data.ID,
		"name":  data.User.Name,
		"email": data.User.Email,
		"role":  data.Role,
	}
}

func OrganizationMemberCollection(data []models.OrganizationMember) []map[string]any {
	return Collect(data, OrganizationMemberResource)
}
