import { getRouteApi } from '@tanstack/react-router'
import { GroupActivityGrid } from '@/features/activities/components/GroupActivityGrid'
import { useActivities } from '@/features/activities/activities-api'
import { DashboardFilters } from '@/features/dashboard/components/DashboardFilters'
import { Leaderboard } from '@/features/dashboard/components/Leaderboard'
import { parseIds, serializeIds, type DashboardFilter } from '@/features/dashboard/filters'
import { useUsers } from '@/features/users/users-api'
import { useGridRange } from '@/lib/use-grid-range'
import { usePageTitle } from '@/lib/use-page-title'

const dashboardRoute = getRouteApi('/_authenticated/')

export function DashboardPage() {
  usePageTitle('Pulpit')
  const search = dashboardRoute.useSearch()
  const navigate = dashboardRoute.useNavigate()
  const users = useUsers()
  const activities = useActivities()
  const { range, setRange, layout } = useGridRange('grid:group')

  const filter: DashboardFilter = { userIds: parseIds(search.osoby), activityIds: parseIds(search.aktywnosci) }
  const setFilter = (next: DashboardFilter) =>
    void navigate({
      search: { osoby: serializeIds(next.userIds), aktywnosci: serializeIds(next.activityIds) },
      replace: true,
    })

  return (
    <div className="flex flex-col gap-5">
      <h2 className="text-[26px] sm:text-[32px]">Pulpit</h2>

      <DashboardFilters
        users={users.data ?? []}
        activities={activities.data ?? []}
        filter={filter}
        onChange={setFilter}
      />

      <div className="grid gap-5 lg:grid-cols-2">
        <GroupActivityGrid range={range} onRangeChange={setRange} layout={layout} filter={filter} size="lg" />
        <Leaderboard from={layout.firstDay} to={layout.lastDay} range={range} onRangeChange={setRange} />
      </div>
    </div>
  )
}
