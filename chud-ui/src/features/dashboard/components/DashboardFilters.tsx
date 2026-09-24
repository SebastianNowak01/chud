import { Link } from '@tanstack/react-router'
import { Card } from '@/components/ui/Card'
import { Hint } from '@/components/ui/Hint'
import { MultiCombobox } from '@/components/ui/MultiCombobox'
import type { Activity } from '@/features/activities/types'
import type { DashboardFilter } from '@/features/dashboard/filters'
import type { User } from '@/features/users/types'

interface DashboardFiltersProps {
  users: User[]
  activities: Activity[]
  filter: DashboardFilter
  onChange: (filter: DashboardFilter) => void
}

export function DashboardFilters({ users, activities, filter, onChange }: DashboardFiltersProps) {
  const active = filter.userIds.length + filter.activityIds.length > 0

  return (
    <Card className="z-10 flex flex-col gap-3">
      <div className="grid items-start gap-4 sm:grid-cols-2">
        <MultiCombobox
          label="Osoby"
          placeholder="Wszyscy"
          emptyText="Nikogo takiego nie ma."
          options={users.map((u) => ({ id: u.id, label: u.username, color: u.color }))}
          selected={filter.userIds}
          onChange={(userIds) => onChange({ ...filter, userIds })}
        />
        <div className="flex flex-col gap-1">
          <MultiCombobox
            label="Aktywności"
            placeholder="Wszystkie"
            emptyText="Nie ma takiej aktywności."
            options={[...activities]
              .sort((a, b) => Number(Boolean(a.archivedAt)) - Number(Boolean(b.archivedAt)))
              .map((a) => ({ id: a.id, label: a.archivedAt ? `${a.name} (archiwum)` : a.name }))}
            selected={filter.activityIds}
            onChange={(activityIds) => onChange({ ...filter, activityIds })}
          />
          {activities.length === 0 && (
            <Hint as="span">
              Nie ma jeszcze aktywności. <Link to="/activities">Dodaj pierwszą</Link>
            </Hint>
          )}
        </div>
      </div>
      {active && (
        <button
          type="button"
          className="cursor-pointer self-start bg-transparent p-0 text-[14px] text-muted underline hover:text-ink"
          onClick={() => onChange({ userIds: [], activityIds: [] })}
        >
          Wyczyść filtry
        </button>
      )}
    </Card>
  )
}
