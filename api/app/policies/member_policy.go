package policies

import (
	"api/app/constants"
	"api/app/models"
)

type MemberPolicy struct{}

func (MemberPolicy) View(m models.OrganizationMember) bool {
	return Allows(m, constants.MembersView)
}

func (MemberPolicy) Invite(m models.OrganizationMember) bool {
	return Allows(m, constants.MembersInvite)
}

func (MemberPolicy) Update(actor, target models.OrganizationMember) bool {
	return Allows(actor, constants.MembersUpdate) && target.Role != constants.OwnerRole
}

func (MemberPolicy) Remove(actor, target models.OrganizationMember) bool {
	return Allows(actor, constants.MembersRemove) && target.Role != constants.OwnerRole
}

func (MemberPolicy) ViewInvites(m models.OrganizationMember) bool {
	return Allows(m, constants.InvitesView)
}

func (MemberPolicy) RevokeInvite(m models.OrganizationMember) bool {
	return Allows(m, constants.InvitesRevoke)
}
