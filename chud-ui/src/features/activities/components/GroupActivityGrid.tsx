import { ContributionGrid } from '@/components/common/ContributionGrid'
import type { CardSize } from '@/components/ui/Card'
import { useActivities } from '@/features/activities/activities-api'
import { filterEntries, type DashboardFilter } from '@/features/dashboard/filters'
import { useEntriesInRange } from '@/features/entries/entries-api'
import { useUsersById } from '@/features/users/users-api'
import type { GridLayout, GridRange } from '@/lib/contributions'

// Everyone's entries in every activity, each day striped in the colors of the people active on it.
interface GroupActivityGridProps {
  range: GridRange
  onRangeChange: (range: GridRange) => void
  layout: GridLayout
  filter: DashboardFilter
  size?: CardSize
}

export function GroupActivityGrid({ range, onRangeChange, layout, filter, size }: GroupActivityGridProps) {
  const entries = useEntriesInRange(layout.firstDay, layout.lastDay)
  const activities = useActivities()
  const usersById = useUsersById()
  const activitiesById = new Map((activities.data ?? []).map((a) => [a.id, a]))
  const visible = entries.data && filterEntries(entries.data, filter)
  const onlyUser = filter.userIds.length === 1 ? usersById.get(filter.userIds[0]) : undefined

  return (
    <ContributionGrid
      title="Aktywność grupy"
      layout={layout}
      range={range}
      onRangeChange={onRangeChange}
      entries={visible}
      error={entries.error}
      onRetry={() => void entries.refetch()}
      coloring={{ kind: 'single', color: onlyUser?.color ?? 'var(--color-ink)' }}
      usersById={usersById}
      activitiesById={activitiesById}
      size={size}
    />
  )
}
