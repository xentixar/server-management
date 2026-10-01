// Package constants defines roles, permissions, and other static application values.
package constants

type Role struct {
	Name        string
	Permissions []Permission
}

const (
	OwnerRole  = "owner"
	AdminRole  = "admin"
	ViewerRole = "viewer"
)

var Roles = map[string]Role{
	OwnerRole: {
		Name:        OwnerRole,
		Permissions: Permissions,
	},
	AdminRole: {
		Name:        AdminRole,
		Permissions: Permissions,
	},
	ViewerRole: {
		Name:        ViewerRole,
		Permissions: Permissions,
	},
}

func IsValidRole(role string) bool {
	_, exists := Roles[role]
	return exists
}
