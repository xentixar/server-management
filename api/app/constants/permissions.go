package constants

type Permission struct {
	Key   string
	Label string
}

const (
	PermissionOrganizationCreate = "organization.create"
	PermissionOrganizationList   = "organization.list"
	PermissionOrganizationShow   = "organization.show"
	PermissionOrganizationUpdate = "organization.update"
	PermissionOrganizationDelete = "organization.delete"
)

var Permissions = []Permission{
	{Key: PermissionOrganizationCreate, Label: "Create Organization"},
	{Key: PermissionOrganizationList, Label: "List Organizations"},
	{Key: PermissionOrganizationShow, Label: "View Organization"},
	{Key: PermissionOrganizationUpdate, Label: "Update Organization"},
	{Key: PermissionOrganizationDelete, Label: "Delete Organization"},
}
