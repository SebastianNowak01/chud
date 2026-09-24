import { DrawablyButton } from 'drawably/react'
import { Card } from '@/components/ui/Card'
import { FormError } from '@/components/ui/FormError'
import { useRestoreActivity } from '@/features/activities/activities-api'
import type { Activity } from '@/features/activities/types'
import { buttonState } from '@/lib/button-state'
import { LOCALE } from '@/lib/dates'

export function ArchivedBanner({ activity, canManage }: { activity: Activity; canManage: boolean }) {
  const restore = useRestoreActivity()
  const since = activity.archivedAt
    ? new Date(activity.archivedAt).toLocaleDateString(LOCALE, { day: 'numeric', month: 'long', year: 'numeric' })
    : ''

  return (
    <Card className="flex flex-col gap-2 [--drawably-stroke:var(--color-excused)]">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="m-0">
          <span className="font-semibold">Ta aktywność jest w archiwum</span>
          {since && <span className="text-muted"> od {since}</span>}. Historia zostaje, ale nic nowego nie da się dodać.
        </p>
        {canManage && (
          <DrawablyButton state={buttonState(restore.status)} onClick={() => restore.mutate(activity.id)}>
            Przywróć
          </DrawablyButton>
        )}
      </div>
      <FormError error={restore.error} />
    </Card>
  )
}
