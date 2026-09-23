import { ContributionGrid } from '@/components/common/ContributionGrid'
import type { Activity } from '@/features/activities/types'
import type { Entry } from '@/features/entries/types'
import { useUsersById } from '@/features/users/users-api'
import { useGridRange } from '@/lib/use-grid-range'

interface ActivityGridProps {
  activity: Activity
  entries: Entry[] | undefined
  error?: Error | null
}

export function ActivityGrid({ activity, entries, error }: ActivityGridProps) {
  const { range, setRange, layout } = useGridRange('grid:activity')
  const usersById = useUsersById()
  const activitiesById = new Map([[activity.id, activity]])
  const inRange = entries?.filter((entry) => {
    const occurredAt = new Date(entry.occurredAt)
    return occurredAt >= layout.from && occurredAt < layout.to
  })

  return (
    <ContributionGrid
      title="Activity"
      layout={layout}
      range={range}
      onRangeChange={setRange}
      entries={inRange}
      error={error}
      coloring={{ kind: 'people' }}
      usersById={usersById}
      activitiesById={activitiesById}
    />
  )
}
