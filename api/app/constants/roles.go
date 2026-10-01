package constants

type Role struct {
	Name        string
	Permissions []Permission
}

var ROLES = []Role{
	{
		Name:        "owner",
		Permissions: PERMISSIONS,
	},
	{
		Name:        "admin",
		Permissions: PERMISSIONS,
	},
	{
		Name:        "viewer",
		Permissions: PERMISSIONS,
	},
}
