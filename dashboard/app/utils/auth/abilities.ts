export const Ability = {
  OrganizationCreate: 'organization.create',
  OrganizationList: 'organization.list',
  OrganizationShow: 'organization.show',
  OrganizationUpdate: 'organization.update',
  OrganizationDelete: 'organization.delete',
  MembersView: 'organization.members.view',
  MembersInvite: 'organization.members.invite',
  MembersUpdate: 'organization.members.update',
  MembersRemove: 'organization.members.remove',
  InvitesView: 'organization.invites.view',
  InvitesRevoke: 'organization.invites.revoke',
  BillingView: 'organization.billing.view',
  BillingManage: 'organization.billing.manage',
  ServersView: 'organization.servers.view',
  ServersCreate: 'organization.servers.create',
  ServersUpdate: 'organization.servers.update',
  ServersDelete: 'organization.servers.delete',
  RuntimesView: 'organization.servers.runtimes.view',
  RuntimesManage: 'organization.servers.runtimes.manage',
  ServicesView: 'organization.servers.services.view',
  ServicesManage: 'organization.servers.services.manage',
  CronsView: 'organization.servers.crons.view',
  CronsManage: 'organization.servers.crons.manage'
} as const

export type Ability = typeof Ability[keyof typeof Ability]
