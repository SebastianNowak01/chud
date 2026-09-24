import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { DrawablyButton } from 'drawably/react'
import { Card } from '@/components/ui/Card'
import { Hint } from '@/components/ui/Hint'
import { LoadError } from '@/components/ui/LoadError'
import { Modal } from '@/components/ui/Modal'
import { ActivityForm } from '@/features/activities/components/ActivityForm'
import { useActivities } from '@/features/activities/activities-api'

export function ActivitiesPage() {
  const activities = useActivities()
  const [creating, setCreating] = useState(false)

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <h2 className="text-[26px] sm:text-[32px]">Aktywności</h2>
        <DrawablyButton variant="solid" onClick={() => setCreating(true)}>
          Nowa aktywność
        </DrawablyButton>
      </div>

      {creating && (
        <Modal title="Nowa aktywność" onClose={() => setCreating(false)}>
          <ActivityForm onDone={() => setCreating(false)} />
        </Modal>
      )}

      {activities.isPending && <Hint>Ładowanie…</Hint>}
      <LoadError error={activities.error} onRetry={() => void activities.refetch()} />
      {activities.data?.length === 0 && <Hint>Nie ma jeszcze aktywności. Utwórz pierwszą.</Hint>}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {activities.data?.map((activity) => (
          <Link
            key={activity.id}
            to="/activities/$activityId"
            params={{ activityId: activity.id }}
            className="text-inherit no-underline"
          >
            <Card size="lg" className="flex h-full flex-col gap-1.5">
              <h3 className="text-[22px]">{activity.name}</h3>
              {activity.description && <p className="text-[15px] text-muted">{activity.description}</p>}
            </Card>
          </Link>
        ))}
      </div>
    </div>
  )
}
