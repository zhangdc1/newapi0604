import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { Save, Settings2 } from 'lucide-react'
import { formatQuota } from '@/lib/format'
import { SectionPageLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  getAdminDistributionRecords,
  getAdminDistributionSettings,
  updateAdminDistributionSettings,
} from './api'

type DistributionSetting = {
  enabled: boolean
  reward_mode: number
  reward_rate: number
  fixed_reward_quota: number
  referral_limit: number
  freeze_days: number
  min_settlement_quota: number
  admin_recharge_trigger: boolean
  redeem_temporary_attribution: boolean
}

type CommissionRecord = {
  id: number
  referrer_user_id: number
  referred_user_id: number
  source: string
  increased_quota: number
  commission_quota: number
  status: number
  created_at: number
}

const DEFAULT_SETTING: DistributionSetting = {
  enabled: false,
  reward_mode: 1,
  reward_rate: 10,
  fixed_reward_quota: 0,
  referral_limit: 0,
  freeze_days: 0,
  min_settlement_quota: 0,
  admin_recharge_trigger: false,
  redeem_temporary_attribution: true,
}

function itemsFromPage<T>(payload: unknown): T[] {
  const data = payload as { data?: { items?: T[] } }
  return data?.data?.items ?? []
}

function numberValue(value: string) {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

export function DistributionAdmin() {
  const [setting, setSetting] = useState<DistributionSetting>(DEFAULT_SETTING)
  const [records, setRecords] = useState<CommissionRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      const [settingRes, recordsRes] = await Promise.all([
        getAdminDistributionSettings(),
        getAdminDistributionRecords({ page_size: 10 }),
      ])
      if (settingRes.success && settingRes.data) {
        setSetting({ ...DEFAULT_SETTING, ...settingRes.data })
      }
      setRecords(itemsFromPage<CommissionRecord>(recordsRes))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const updateField = <K extends keyof DistributionSetting>(
    key: K,
    value: DistributionSetting[K]
  ) => {
    setSetting((prev) => ({ ...prev, [key]: value }))
  }

  const save = async () => {
    setSaving(true)
    try {
      const res = await updateAdminDistributionSettings(setting)
      if (res.success) {
        toast.success('Distribution settings saved')
        await load()
      }
    } finally {
      setSaving(false)
    }
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>Distribution Management</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='mx-auto grid w-full max-w-7xl gap-4 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]'>
          <Card>
            <CardHeader>
              <CardTitle className='flex items-center gap-2 text-base'>
                <Settings2 className='size-4' />
                Settings
              </CardTitle>
            </CardHeader>
            <CardContent className='grid gap-4'>
              <div className='flex items-center gap-2'>
                <Checkbox
                  id='distribution-enabled'
                  checked={setting.enabled}
                  onCheckedChange={(checked) =>
                    updateField('enabled', checked === true)
                  }
                />
                <Label htmlFor='distribution-enabled'>Enable distribution</Label>
              </div>

              <div className='grid gap-2'>
                <Label>Reward mode</Label>
                <Select
                  value={String(setting.reward_mode)}
                  onValueChange={(value) =>
                    updateField('reward_mode', Number(value))
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value='1'>Percentage</SelectItem>
                    <SelectItem value='2'>Fixed quota</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className='grid gap-3 md:grid-cols-2'>
                <div className='grid gap-2'>
                  <Label>Reward rate (%)</Label>
                  <Input
                    type='number'
                    min={0}
                    max={50}
                    value={setting.reward_rate}
                    onChange={(e) =>
                      updateField('reward_rate', numberValue(e.target.value))
                    }
                  />
                </div>
                <div className='grid gap-2'>
                  <Label>Fixed reward quota</Label>
                  <Input
                    type='number'
                    min={0}
                    value={setting.fixed_reward_quota}
                    onChange={(e) =>
                      updateField(
                        'fixed_reward_quota',
                        numberValue(e.target.value)
                      )
                    }
                  />
                </div>
                <div className='grid gap-2'>
                  <Label>Reward topup limit</Label>
                  <Input
                    type='number'
                    min={0}
                    value={setting.referral_limit}
                    onChange={(e) =>
                      updateField('referral_limit', numberValue(e.target.value))
                    }
                  />
                </div>
                <div className='grid gap-2'>
                  <Label>Freeze days</Label>
                  <Input
                    type='number'
                    min={0}
                    value={setting.freeze_days}
                    onChange={(e) =>
                      updateField('freeze_days', numberValue(e.target.value))
                    }
                  />
                </div>
                <div className='grid gap-2 md:col-span-2'>
                  <Label>Minimum settlement quota</Label>
                  <Input
                    type='number'
                    min={0}
                    value={setting.min_settlement_quota}
                    onChange={(e) =>
                      updateField(
                        'min_settlement_quota',
                        numberValue(e.target.value)
                      )
                    }
                  />
                </div>
              </div>

              <div className='grid gap-3'>
                <div className='flex items-center gap-2'>
                  <Checkbox
                    id='admin-recharge-trigger'
                    checked={setting.admin_recharge_trigger}
                    onCheckedChange={(checked) =>
                      updateField('admin_recharge_trigger', checked === true)
                    }
                  />
                  <Label htmlFor='admin-recharge-trigger'>
                    Reward admin balance additions
                  </Label>
                </div>
                <div className='flex items-center gap-2'>
                  <Checkbox
                    id='redeem-temporary-attribution'
                    checked={setting.redeem_temporary_attribution}
                    onCheckedChange={(checked) =>
                      updateField(
                        'redeem_temporary_attribution',
                        checked === true
                      )
                    }
                  />
                  <Label htmlFor='redeem-temporary-attribution'>
                    Allow redemption temporary attribution
                  </Label>
                </div>
              </div>

              <Button onClick={save} disabled={saving || loading}>
                <Save className='size-4' />
                Save
              </Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className='text-base'>Recent Commission Records</CardTitle>
            </CardHeader>
            <CardContent className='space-y-2'>
              {records.length === 0 ? (
                <div className='text-muted-foreground text-sm'>No records</div>
              ) : (
                records.map((record) => (
                  <div
                    key={record.id}
                    className='border-border grid gap-2 border-b py-3 text-sm last:border-0 md:grid-cols-[1fr_auto]'
                  >
                    <div>
                      <div className='font-medium'>
                        Referrer #{record.referrer_user_id} / Referred #
                        {record.referred_user_id}
                      </div>
                      <div className='text-muted-foreground text-xs'>
                        {record.source} · status {record.status}
                      </div>
                    </div>
                    <div className='text-left md:text-right'>
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
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
