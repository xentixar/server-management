package resources

import "api/app/models"

func Organization(data models.Organization) map[string]any {
	return map[string]any{
		"id":   data.ID,
		"name": data.Name,
		"slug": data.Slug,
	}
}
