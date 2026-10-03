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

func (ServerPolicy) ViewRuntimes(m models.OrganizationMember) bool {
	return Allows(m, constants.RuntimesView)
}

func (ServerPolicy) ManageRuntimes(m models.OrganizationMember) bool {
	return Allows(m, constants.RuntimesManage)
}

func (ServerPolicy) ViewServices(m models.OrganizationMember) bool {
	return Allows(m, constants.ServicesView)
}

func (ServerPolicy) ManageServices(m models.OrganizationMember) bool {
	return Allows(m, constants.ServicesManage)
}

func (ServerPolicy) ViewCrons(m models.OrganizationMember) bool {
	return Allows(m, constants.CronsView)
}

func (ServerPolicy) ManageCrons(m models.OrganizationMember) bool {
	return Allows(m, constants.CronsManage)
}

func (ServerPolicy) ViewFirewall(m models.OrganizationMember) bool {
	return Allows(m, constants.FirewallView)
}

func (ServerPolicy) ManageFirewall(m models.OrganizationMember) bool {
	return Allows(m, constants.FirewallManage)
}
