import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { DrawablyButton } from 'drawably/react'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { Modal } from '@/components/ui/Modal'
import { ActivityForm } from '@/features/activities/components/ActivityForm'
import { GroupActivityGrid } from '@/features/activities/components/GroupActivityGrid'
import { useActivities } from '@/features/activities/activities-api'
import { Leaderboard } from '@/features/dashboard/components/Leaderboard'
import { useGridRange } from '@/lib/use-grid-range'

export function ActivitiesPage() {
  const activities = useActivities()
  const { range, setRange, layout } = useGridRange('grid:group')
  const [creating, setCreating] = useState(false)

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <h2 className="text-[26px] sm:text-[32px]">Pulpit</h2>
        <DrawablyButton variant="solid" onClick={() => setCreating(true)}>
          Nowa aktywność
        </DrawablyButton>
      </div>

      {creating && (
        <Modal title="Nowa aktywność" onClose={() => setCreating(false)}>
          <ActivityForm onDone={() => setCreating(false)} />
        </Modal>
      )}

      <div className="grid gap-5 lg:grid-cols-2">
        <GroupActivityGrid range={range} onRangeChange={setRange} layout={layout} size="lg" />
        <Leaderboard from={layout.firstDay} to={layout.lastDay} range={range} onRangeChange={setRange} />
      </div>

      <h3 className="text-[22px]">Aktywności</h3>
      {activities.isPending && <p className="text-[15px] text-muted">Ładowanie…</p>}
      {activities.error && <ErrorText>{activities.error.message}</ErrorText>}
      {activities.data?.length === 0 && (
        <p className="text-[15px] text-muted">Nie ma jeszcze aktywności. Utwórz pierwszą.</p>
      )}

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
