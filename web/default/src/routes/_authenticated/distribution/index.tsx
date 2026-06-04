import { createFileRoute } from '@tanstack/react-router'
import { DistributionCenter } from '@/features/distribution'

export const Route = createFileRoute('/_authenticated/distribution/')({
  component: DistributionCenter,
})
