package policies

import (
	"api/app/constants"
	"api/app/models"
)

// OrganizationPolicy — add methods as new org abilities appear.
type OrganizationPolicy struct{}

func (OrganizationPolicy) Create(_ models.User) bool {
	return true
}

func (OrganizationPolicy) View(m models.OrganizationMember) bool {
	return Allows(m, constants.OrganizationShow)
}

func (OrganizationPolicy) Update(m models.OrganizationMember) bool {
	return Allows(m, constants.OrganizationUpdate)
}

func (OrganizationPolicy) Delete(m models.OrganizationMember) bool {
	return Allows(m, constants.OrganizationDelete)
}

func (OrganizationPolicy) ViewBilling(m models.OrganizationMember) bool {
	return Allows(m, constants.BillingView)
}

func (OrganizationPolicy) ManageBilling(m models.OrganizationMember) bool {
	return Allows(m, constants.BillingManage)
}
