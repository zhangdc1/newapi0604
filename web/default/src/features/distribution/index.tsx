import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { Copy, Gift, LinkIcon, Users } from 'lucide-react'
import { formatQuota } from '@/lib/format'
import { SectionPageLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  getDistributionInvites,
  getDistributionOverview,
  getDistributionRecords,
  settleDistribution,
} from './api'

type Overview = {
  invite_link: string
  invite_count: number
  effective_invites: number
  pending_quota: number
  available_quota: number
  settled_quota: number
  total_earned_quota: number
  min_settlement_quota: number
  enabled: boolean
}

type CommissionRecord = {
  id: number
  referred_user_id: number
  source: string
  increased_quota: number
  commission_quota: number
  status: number
  freeze_until: number
  created_at: number
}

type InviteRecord = {
  id: number
  username: string
  total_increased_quota: number
  total_commission_quota: number
  reward_count: number
}

function itemsFromPage<T>(payload: unknown): T[] {
  const data = payload as { data?: { items?: T[] } }
  return data?.data?.items ?? []
}

function statusText(status: number) {
  if (status === 1) return 'Pending'
  if (status === 2) return 'Settled'
  if (status === 3) return 'Invalid'
  return 'Unknown'
}

export function DistributionCenter() {
  const [overview, setOverview] = useState<Overview | null>(null)
  const [records, setRecords] = useState<CommissionRecord[]>([])
  const [invites, setInvites] = useState<InviteRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [settling, setSettling] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      const [overviewRes, recordsRes, invitesRes] = await Promise.all([
        getDistributionOverview(),
        getDistributionRecords(),
        getDistributionInvites(),
      ])
      if (overviewRes.success) setOverview(overviewRes.data)
      setRecords(itemsFromPage<CommissionRecord>(recordsRes))
      setInvites(itemsFromPage<InviteRecord>(invitesRes))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const copyInviteLink = async () => {
    if (!overview?.invite_link) return
    await navigator.clipboard.writeText(overview.invite_link)
    toast.success('Invite link copied')
  }

  const handleSettle = async () => {
    setSettling(true)
    try {
      const res = await settleDistribution()
      if (res.success) {
        toast.success('Commission transferred to balance')
        await load()
      }
    } finally {
      setSettling(false)
    }
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>Distribution</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='mx-auto flex w-full max-w-7xl flex-col gap-4'>
          <Card>
            <CardHeader>
              <CardTitle className='flex items-center gap-2 text-base'>
                <LinkIcon className='size-4' />
                Invite Link
              </CardTitle>
            </CardHeader>
            <CardContent className='flex flex-col gap-3 sm:flex-row'>
              <Input
                value={overview?.invite_link ?? ''}
                readOnly
                className='font-mono text-xs'
              />
              <Button onClick={copyInviteLink} disabled={!overview?.invite_link}>
                <Copy className='size-4' />
                Copy
              </Button>
            </CardContent>
          </Card>

          <div className='grid gap-3 md:grid-cols-4'>
            {[
              ['Invites', overview?.invite_count ?? 0, Users],
              ['Effective', overview?.effective_invites ?? 0, Users],
              ['Available', formatQuota(overview?.available_quota ?? 0), Gift],
              ['Total Earned', formatQuota(overview?.total_earned_quota ?? 0), Gift],
            ].map(([label, value, Icon]) => (
              <Card key={String(label)}>
                <CardContent className='flex items-center justify-between p-4'>
                  <div>
                    <div className='text-muted-foreground text-xs'>{label}</div>
                    <div className='mt-1 text-lg font-semibold'>{String(value)}</div>
                  </div>
                  <Icon className='text-muted-foreground size-5' />
                </CardContent>
              </Card>
            ))}
          </div>

          <Card>
            <CardHeader>
              <CardTitle className='text-base'>Settlement</CardTitle>
            </CardHeader>
            <CardContent className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
              <div className='text-sm'>
                Available commission:{' '}
                <span className='font-semibold'>
                  {formatQuota(overview?.available_quota ?? 0)}
                </span>
                <span className='text-muted-foreground ml-3'>
                  Minimum: {formatQuota(overview?.min_settlement_quota ?? 0)}
                </span>
              </div>
              <Button
                onClick={handleSettle}
                disabled={settling || loading || (overview?.available_quota ?? 0) <= 0}
              >
                Transfer to Balance
              </Button>
            </CardContent>
          </Card>

          <div className='grid gap-4 xl:grid-cols-2'>
            <Card>
              <CardHeader>
                <CardTitle className='text-base'>Recent Rewards</CardTitle>
              </CardHeader>
              <CardContent className='space-y-2'>
                {records.length === 0 ? (
                  <div className='text-muted-foreground text-sm'>No records</div>
                ) : (
                  records.map((record) => (
                    <div
                      key={record.id}
                      className='border-border flex items-center justify-between border-b py-2 text-sm last:border-0'
                    >
                      <div>
                        <div className='font-medium'>
                          User #{record.referred_user_id} · {record.source}
                        </div>
                        <div className='text-muted-foreground text-xs'>
                          {statusText(record.status)}
                        </div>
                      </div>
                      <div className='text-right'>
                        <div>{formatQuota(record.commission_quota)}</div>
                        <div className='text-muted-foreground text-xs'>
                          from {formatQuota(record.increased_quota)}
                        </div>
                      </div>
                    </div>
                  ))
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className='text-base'>Invited Users</CardTitle>
              </CardHeader>
              <CardContent className='space-y-2'>
                {invites.length === 0 ? (
                  <div className='text-muted-foreground text-sm'>No invites</div>
                ) : (
                  invites.map((invite) => (
                    <div
                      key={invite.id}
                      className='border-border flex items-center justify-between border-b py-2 text-sm last:border-0'
                    >
                      <div>
                        <div className='font-medium'>{invite.username}</div>
                        <div className='text-muted-foreground text-xs'>
                          {invite.reward_count} rewarded topups
                        </div>
                      </div>
                      <div className='text-right'>
                        <div>{formatQuota(invite.total_commission_quota)}</div>
                        <div className='text-muted-foreground text-xs'>
                          contributed {formatQuota(invite.total_increased_quota)}
                        </div>
                      </div>
                    </div>
                  ))
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
