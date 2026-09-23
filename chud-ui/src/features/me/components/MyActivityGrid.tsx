import { ContributionGrid } from '@/components/common/ContributionGrid'
import { useActivities } from '@/features/activities/activities-api'
import { useMyEntriesInRange } from '@/features/entries/entries-api'
import type { User } from '@/features/users/types'
import { useGridRange } from '@/lib/use-grid-range'

// My entries in every activity, in my color (or the color being picked, so the preview is live).
export function MyActivityGrid({ me, color }: { me: User; color: string }) {
  const { range, setRange, layout } = useGridRange('grid:me')
  const entries = useMyEntriesInRange(layout.firstDay, layout.lastDay)
  const activities = useActivities()
  const usersById = new Map([[me.id, { ...me, color }]])
  const activitiesById = new Map((activities.data ?? []).map((a) => [a.id, a]))

  return (
    <ContributionGrid
      title="Moja aktywność"
      layout={layout}
      range={range}
      onRangeChange={setRange}
      entries={entries.data}
      error={entries.error}
      coloring={{ kind: 'single', color }}
      usersById={usersById}
      activitiesById={activitiesById}
    />
  )
}
