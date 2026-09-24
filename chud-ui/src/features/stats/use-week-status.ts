import { useMemo, useState } from 'react'
import type { WeekStatus } from '@/components/common/ContributionGrid'
import { useUserEntriesInRange } from '@/features/entries/entries-api'
import { weeklyCounts } from '@/features/stats/mood'
import { useOccurrences } from '@/features/stats/stats-api'
import { currentWeek, toDateString, weekStart } from '@/lib/dates'

const useWeekCounts = (userId: string, from: string) => {
  const occurrences = useOccurrences(from, currentWeek().to)
  const entries = useUserEntriesInRange(userId, from, toDateString(new Date()))
  const weeks = useMemo(
    () => weeklyCounts(occurrences.data ?? [], entries.data ?? [], userId),
    [occurrences.data, entries.data, userId],
  )
  return {
    weeks,
    loaded: occurrences.data !== undefined && entries.data !== undefined,
    error: occurrences.error ?? entries.error,
    retry: () => {
      void occurrences.refetch()
      void entries.refetch()
    },
  }
}

export const useWeekStatus = (userId: string, firstDay: string) => {
  const [focusedDate, setFocusedDate] = useState(() => toDateString(new Date()))
  const { weeks, loaded, error, retry } = useWeekCounts(userId, firstDay)
  const week = weekStart(focusedDate)

  const gridStatus = useMemo<WeekStatus>(() => ({ onFocusDate: setFocusedDate }), [])

  return { week, counts: weeks.get(week), loaded, error, retry, gridStatus }
}
