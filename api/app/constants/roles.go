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

var ROLES = map[string]Role{
	OwnerRole: {
		Name:        "owner",
		Permissions: PERMISSIONS,
	},
	AdminRole: {
		Name:        "admin",
		Permissions: PERMISSIONS,
	},
	ViewerRole: {
		Name:        "viewer",
		Permissions: PERMISSIONS,
	},
}

func IsValidRole(role string) bool {
	_, exists := ROLES[role]
	return exists
}
