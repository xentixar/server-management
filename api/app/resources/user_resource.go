// Package resources maps domain models to API response payloads.
package resources

import "api/app/models"

func UserResource(data models.User) map[string]any {
	return map[string]any{
		"id":    data.ID,
		"name":  data.Name,
		"email": data.Email,
	}
}

func UserCollection(data []models.User) []map[string]any {
	return Collect(data, UserResource)
}
