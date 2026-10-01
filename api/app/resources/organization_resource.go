package resources

import (
	"api/app/constants"
	"api/app/models"
)

func OrganizationResource(data models.Organization, role string) map[string]any {
	return map[string]any{
		"id":          data.ID,
		"name":        data.Name,
		"slug":        data.Slug,
		"role":        role,
		"permissions": constants.PermissionKeysForRole(role),
	}
}

func OrganizationCollection(data []models.Organization, rolesByOrgID map[uint]string) []map[string]any {
	items := make([]map[string]any, 0, len(data))
	for _, organization := range data {
		items = append(items, OrganizationResource(organization, rolesByOrgID[organization.ID]))
	}
	return items
}
