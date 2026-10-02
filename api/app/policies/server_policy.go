package policies

import (
	"api/app/constants"
	"api/app/models"
)

type ServerPolicy struct{}

func (ServerPolicy) View(m models.OrganizationMember) bool {
	return Allows(m, constants.ServersView)
}

func (ServerPolicy) Create(m models.OrganizationMember) bool {
	return Allows(m, constants.ServersCreate)
}

func (ServerPolicy) Update(m models.OrganizationMember) bool {
	return Allows(m, constants.ServersUpdate)
}

func (ServerPolicy) Delete(m models.OrganizationMember) bool {
	return Allows(m, constants.ServersDelete)
}
