import type { BadgeTone } from '@/components/ui/Badge'
import type { OccurrenceStatus } from '@/features/stats/types'

export const STATUS_LABEL: Record<OccurrenceStatus, string> = {
  done: '✓ zrobione',
  excused: 'wymówka',
  missed: '✗ opuszczone',
  pending: 'do zrobienia',
}

export const STATUS_TONE: Record<OccurrenceStatus, BadgeTone> = {
  done: 'done',
  excused: 'excused',
  missed: 'danger',
  pending: 'inherit',
}
