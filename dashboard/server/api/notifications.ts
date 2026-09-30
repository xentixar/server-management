import { sub } from 'date-fns'

// ponytail: mock system events until API /notifications lands
const notifications = [{
  id: 1,
  unread: true,
  title: 'Server provisioned',
  body: 'edge-01 is ready.',
  date: sub(new Date(), { minutes: 12 }).toISOString()
}, {
  id: 2,
  unread: true,
  title: 'Deploy finished',
  body: 'api.example.com deployed successfully.',
  date: sub(new Date(), { hours: 2 }).toISOString()
}, {
  id: 3,
  title: 'Member invited',
  body: 'ada@example.com was invited to Acme.',
  date: sub(new Date(), { days: 1 }).toISOString()
}, {
  id: 4,
  title: 'Certificate expiring',
  body: 'app.example.com expires in 14 days.',
  date: sub(new Date(), { days: 3 }).toISOString()
}]

export default eventHandler(async () => {
  return notifications
})
