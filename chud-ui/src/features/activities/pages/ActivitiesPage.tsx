import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { DrawablyButton } from 'drawably/react'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { ActivityForm } from '@/features/activities/components/ActivityForm'
import { GroupActivityGrid } from '@/features/activities/components/GroupActivityGrid'
import { useActivities } from '@/features/activities/activities-api'

export function ActivitiesPage() {
  const activities = useActivities()
  const [creating, setCreating] = useState(false)

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <h2 className="text-[26px] sm:text-[32px]">Activities</h2>
        {!creating && (
          <DrawablyButton variant="solid" onClick={() => setCreating(true)}>
            New activity
          </DrawablyButton>
        )}
      </div>

      {creating && <ActivityForm onDone={() => setCreating(false)} />}

      <GroupActivityGrid size="lg" />

      {activities.isPending && <p className="text-[15px] text-muted">Loading…</p>}
      {activities.error && <ErrorText>{activities.error.message}</ErrorText>}
      {activities.data?.length === 0 && <p className="text-[15px] text-muted">No activities yet. Create the first one.</p>}

      <div className="flex flex-col gap-4">
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
