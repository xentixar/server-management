// ponytail: fixed dates so SSR payload is stable; replace when API lands
const notifications = [{
  id: 1,
  unread: true,
  title: 'Server provisioned',
  body: 'edge-01 is ready.',
  date: '2026-09-30T11:48:00.000Z'
}, {
  id: 2,
  unread: true,
  title: 'Deploy finished',
  body: 'api.example.com deployed successfully.',
  date: '2026-09-30T10:00:00.000Z'
}, {
  id: 3,
  title: 'Member invited',
  body: 'ada@example.com was invited to Acme.',
  date: '2026-09-29T12:00:00.000Z'
}, {
  id: 4,
  title: 'Certificate expiring',
  body: 'app.example.com expires in 14 days.',
  date: '2026-09-27T12:00:00.000Z'
}]

export default eventHandler(async () => {
  return notifications
})
