import { ContributionGrid, type WeekStatus } from '@/components/common/ContributionGrid'
import { useActivities } from '@/features/activities/activities-api'
import { useUserEntriesInRange } from '@/features/entries/entries-api'
import type { User } from '@/features/users/types'
import type { GridLayout, GridRange } from '@/lib/contributions'

interface UserActivityGridProps {
  user: User
  range: GridRange
  onRangeChange: (range: GridRange) => void
  layout: GridLayout
  weekStatus: WeekStatus
}

export function UserActivityGrid({ user, range, onRangeChange, layout, weekStatus }: UserActivityGridProps) {
  const entries = useUserEntriesInRange(user.id, layout.firstDay, layout.lastDay)
  const activities = useActivities()
  const activitiesById = new Map((activities.data ?? []).map((a) => [a.id, a]))

  return (
    <ContributionGrid
      title="Aktywność"
      layout={layout}
      range={range}
      onRangeChange={onRangeChange}
      entries={entries.data}
      error={entries.error}
      coloring={{ kind: 'single', color: user.color }}
      usersById={new Map([[user.id, user]])}
      activitiesById={activitiesById}
      size="lg"
      weekStatus={weekStatus}
    />
  )
}
