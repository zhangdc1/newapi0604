import { api } from '@/lib/api'

export async function getDistributionOverview() {
  const res = await api.get('/api/user/distribution/overview')
  return res.data
}

export async function getDistributionRecords() {
  const res = await api.get('/api/user/distribution/commission-records')
  return res.data
}

export async function getDistributionInvites() {
  const res = await api.get('/api/user/distribution/invites')
  return res.data
}

export async function settleDistribution() {
  const res = await api.post('/api/user/distribution/settle')
  return res.data
}

export async function getDistributionSettings() {
  const res = await api.get('/api/distribution/admin/settings')
  return res.data
}

export async function getAdminDistributionSettings() {
  const res = await api.get('/api/distribution/admin/settings')
  return res.data
}

export async function updateAdminDistributionSettings(data: unknown) {
  const res = await api.put('/api/distribution/admin/settings', data)
  return res.data
}

export async function getAdminDistributionRecords(params?: Record<string, unknown>) {
  const res = await api.get('/api/distribution/admin/commission-records', {
    params,
  })
  return res.data
}

export async function getAdminDistributionTransfers(params?: Record<string, unknown>) {
  const res = await api.get('/api/distribution/admin/transfers', {
    params,
  })
  return res.data
}
