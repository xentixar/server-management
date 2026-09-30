// Package resources contains all the resource mappings
package resources

import "api/app/models"

func UserResource(data models.User) map[string]any {
	return map[string]any{
		"id":    data.ID,
		"name":  data.Name,
		"email": data.Email,
	}
}
