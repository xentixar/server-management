package policies

import (
	"api/app/constants"
	"api/app/models"
)

func Check(role, ability string) bool {
	return constants.RoleHasPermission(role, ability)
}

func Allows(member models.OrganizationMember, ability string) bool {
	return Check(member.Role, ability)
}
