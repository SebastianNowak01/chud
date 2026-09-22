import { ContributionGrid } from '@/components/common/ContributionGrid'
import { useActivities } from '@/features/activities/activities-api'
import { useEntriesInRange } from '@/features/entries/entries-api'
import { useUsersById } from '@/features/users/users-api'
import { useGridRange } from '@/lib/use-grid-range'

// Everyone's entries in every activity, each day striped in the colors of the people active on it.
export function GroupActivityGrid() {
  const { range, setRange, layout } = useGridRange('grid:group')
  const entries = useEntriesInRange(layout.from, layout.to)
  const activities = useActivities()
  const usersById = useUsersById()
  const activitiesById = new Map((activities.data ?? []).map((a) => [a.id, a]))

  return (
    <ContributionGrid
      title="Group activity"
      layout={layout}
      range={range}
      onRangeChange={setRange}
      entries={entries.data}
      error={entries.error}
      coloring={{ kind: 'people' }}
      usersById={usersById}
      activitiesById={activitiesById}
    />
  )
}
