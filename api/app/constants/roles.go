package constants

type Role struct {
	Name        string
	Permissions []string
}

const (
	OwnerRole  = "owner"
	AdminRole  = "admin"
	ViewerRole = "viewer"
)

var RolePermissions = map[string][]string{
	OwnerRole: AllAbilities,
	AdminRole: {
		OrganizationList,
		OrganizationShow,
		OrganizationUpdate,
		MembersView,
		MembersInvite,
		MembersUpdate,
		MembersRemove,
		InvitesView,
		InvitesRevoke,
		BillingView,
		BillingManage,
	},
	ViewerRole: {
		OrganizationList,
		OrganizationShow,
		MembersView,
		InvitesView,
		BillingView,
	},
}

var Roles = map[string]Role{
	OwnerRole:  {Name: OwnerRole, Permissions: RolePermissions[OwnerRole]},
	AdminRole:  {Name: AdminRole, Permissions: RolePermissions[AdminRole]},
	ViewerRole: {Name: ViewerRole, Permissions: RolePermissions[ViewerRole]},
}

func IsValidRole(role string) bool {
	_, ok := Roles[role]
	return ok
}

func RoleHasPermission(roleName, ability string) bool {
	for _, key := range RolePermissions[roleName] {
		if key == ability {
			return true
		}
	}
	return false
}

func PermissionKeysForRole(roleName string) []string {
	keys := RolePermissions[roleName]
	if keys == nil {
		return []string{}
	}
	out := make([]string, len(keys))
	copy(out, keys)
	return out
}
