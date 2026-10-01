package constants

const (
	OrganizationCreate = "organization.create"
	OrganizationList   = "organization.list"
	OrganizationShow   = "organization.show"
	OrganizationUpdate = "organization.update"
	OrganizationDelete = "organization.delete"

	MembersView   = "organization.members.view"
	MembersInvite = "organization.members.invite"
	MembersUpdate = "organization.members.update"
	MembersRemove = "organization.members.remove"

	InvitesView   = "organization.invites.view"
	InvitesRevoke = "organization.invites.revoke"

	BillingView   = "organization.billing.view"
	BillingManage = "organization.billing.manage"
)

var AllAbilities = []string{
	OrganizationCreate,
	OrganizationList,
	OrganizationShow,
	OrganizationUpdate,
	OrganizationDelete,
	MembersView,
	MembersInvite,
	MembersUpdate,
	MembersRemove,
	InvitesView,
	InvitesRevoke,
	BillingView,
	BillingManage,
}
