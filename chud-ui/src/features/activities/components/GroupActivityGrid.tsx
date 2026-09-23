import { ContributionGrid } from '@/components/common/ContributionGrid'
import type { CardSize } from '@/components/ui/Card'
import { useActivities } from '@/features/activities/activities-api'
import { useEntriesInRange } from '@/features/entries/entries-api'
import { useUsersById } from '@/features/users/users-api'
import type { GridLayout, GridRange } from '@/lib/contributions'

// Everyone's entries in every activity, each day striped in the colors of the people active on it.
interface GroupActivityGridProps {
  range: GridRange
  onRangeChange: (range: GridRange) => void
  layout: GridLayout
  size?: CardSize
}

export function GroupActivityGrid({ range, onRangeChange, layout, size }: GroupActivityGridProps) {
  const entries = useEntriesInRange(layout.firstDay, layout.lastDay)
  const activities = useActivities()
  const usersById = useUsersById()
  const activitiesById = new Map((activities.data ?? []).map((a) => [a.id, a]))

  return (
    <ContributionGrid
      title="Aktywność grupy"
      layout={layout}
      range={range}
      onRangeChange={onRangeChange}
      entries={entries.data}
      error={entries.error}
      coloring={{ kind: 'single', color: 'var(--color-ink)' }}
      usersById={usersById}
      activitiesById={activitiesById}
      size={size}
    />
  )
}
