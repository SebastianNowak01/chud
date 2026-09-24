import type { Entry } from '@/features/entries/types'
import type { Occurrence, UserStats } from '@/features/stats/types'
import { addDays, LOCALE, parseDate, toDateString, weekStart } from '@/lib/dates'

export type WeekCounts = Pick<UserStats, 'done' | 'extra' | 'excused' | 'missed' | 'pending'>

export const EMPTY_WEEK: WeekCounts = { done: 0, extra: 0, excused: 0, missed: 0, pending: 0 }

export const FACE_COUNT = 9

const EXTRA_BONUS = 0.05
const MAX_EXTRA_BONUS = 0.25

export const weekScore = ({ done, extra, excused, missed, pending }: WeekCounts): number => {
  const planned = done + excused + missed + pending
  const base = planned === 0 ? 0 : (done - excused - 1.5 * missed) / planned
  const score = base + Math.min(extra * EXTRA_BONUS, MAX_EXTRA_BONUS)
  return Math.max(-1, Math.min(1, score))
}

export const weekFace = (score: number): number => Math.round(((score + 1) / 2) * (FACE_COUNT - 1))

export const weeklyCounts = (occurrences: Occurrence[], entries: Entry[], userId: string): Map<string, WeekCounts> => {
  const weeks = new Map<string, WeekCounts>()
  const countIn = (date: string) => {
    const week = weekStart(date)
    const counts = weeks.get(week) ?? { ...EMPTY_WEEK }
    weeks.set(week, counts)
    return counts
  }
  for (const occurrence of occurrences) {
    if (occurrence.userId === userId) {
      countIn(occurrence.date)[occurrence.status]++
    }
  }
  for (const entry of entries) {
    if (entry.userId === userId && entry.planId === null && !entry.excused) {
      countIn(toDateString(new Date(entry.occurredAt))).extra++
    }
  }
  return weeks
}

const dayMonth = (date: Date) => date.toLocaleDateString(LOCALE, { day: 'numeric', month: 'short' })

export const weekHeading = (week: string, currentWeek: string): string => {
  if (week === currentWeek) {
    return 'Status tygodnia'
  }
  if (week === toDateString(addDays(parseDate(currentWeek), -7))) {
    return 'Status z zeszłego tygodnia'
  }
  const monday = parseDate(week)
  const sunday = addDays(monday, 6)
  const range =
    monday.getMonth() === sunday.getMonth()
      ? `${monday.getDate()}–${dayMonth(sunday)}`
      : `${dayMonth(monday)} – ${dayMonth(sunday)}`
  return `Status tygodnia ${range}`
}

