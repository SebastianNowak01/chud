import type { WeekSummary } from '@/features/summaries/types'
import { addDays, LOCALE, parseDate, toDateString } from '@/lib/dates'

const dayMonth = (date: Date) => date.toLocaleDateString(LOCALE, { day: 'numeric', month: 'short' })

export const weekLabel = (week: string, currentWeek: string): string => {
  if (week === currentWeek) return 'Ten tydzień'
  if (week === toDateString(addDays(parseDate(currentWeek), -7))) return 'Zeszły tydzień'
  const monday = parseDate(week)
  const sunday = addDays(monday, 6)
  return monday.getMonth() === sunday.getMonth()
    ? `${monday.getDate()}–${dayMonth(sunday)}`
    : `${dayMonth(monday)} – ${dayMonth(sunday)}`
}

export const shiftWeek = (week: string, weeks: number): string => toDateString(addDays(parseDate(week), weeks * 7))

export const regenerateWaitMs = (summary: WeekSummary, now: number): number => {
  if (!summary.regenerateAvailableAt) return 0
  return Math.max(0, new Date(summary.regenerateAvailableAt).getTime() - now)
}

export const canRegenerate = (summary: WeekSummary, now: number): boolean =>
  summary.status !== 'generating' &&
  summary.status !== 'disabled' &&
  summary.status !== 'empty' &&
  regenerateWaitMs(summary, now) === 0
