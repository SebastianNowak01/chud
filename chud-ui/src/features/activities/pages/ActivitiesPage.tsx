import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { DrawablyButton, DrawablyCard } from 'drawably/react'
import { ActivityForm } from '@/features/activities/components/ActivityForm'
import { GroupActivityGrid } from '@/features/activities/components/GroupActivityGrid'
import { useActivities } from '@/features/activities/activities-api'

export function ActivitiesPage() {
  const activities = useActivities()
  const [creating, setCreating] = useState(false)

  return (
    <div className="stack">
      <div className="header">
        <h2>Activities</h2>
        {!creating && (
          <DrawablyButton variant="solid" onClick={() => setCreating(true)}>
            New activity
          </DrawablyButton>
        )}
      </div>

      {creating && <ActivityForm onDone={() => setCreating(false)} />}

      <GroupActivityGrid />

      {activities.isPending && <p className="hint">Loading…</p>}
      {activities.error && <p className="error">{activities.error.message}</p>}
      {activities.data?.length === 0 && <p className="hint">No activities yet. Create the first one.</p>}

      <div className="grid">
        {activities.data?.map((activity) => (
          <Link
            key={activity.id}
            to="/activities/$activityId"
            params={{ activityId: activity.id }}
            className="plain-link"
          >
            <DrawablyCard className="card">
              <h3>{activity.name}</h3>
              {activity.description && <p className="hint">{activity.description}</p>}
            </DrawablyCard>
          </Link>
        ))}
      </div>
    </div>
  )
}
