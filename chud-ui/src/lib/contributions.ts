import type { Entry } from '@/features/entries/types'
import { addDays, toDateString } from '@/lib/dates'

export type GridRange = 'month' | 'quarter' | 'year'

export const GRID_RANGES: { value: GridRange; label: string; months: number }[] = [
  { value: 'month', label: 'Miesiąc', months: 1 },
  { value: 'quarter', label: '3 miesiące', months: 3 },
  { value: 'year', label: 'Rok', months: 12 },
]

// What one user did in one activity on one day.
export interface DayItem {
  userId: string
  activityId: string
  done: number
  excused: number
}

export interface Day {
  date: string // YYYY-MM-DD, local
  items: DayItem[]
  entries: Entry[]
  done: number
  excused: number
}

export interface GridLayout {
  from: Date // local midnight of the first Monday
  to: Date // local midnight after today
  firstDay: string // YYYY-MM-DD of from
  lastDay: string // YYYY-MM-DD of today
  weeks: string[][] // columns of 7 dates, Monday first
}

const startOfDay = (date: Date) => new Date(date.getFullYear(), date.getMonth(), date.getDate())

// Weeks covering the range, from the Monday on or before (today - range) up to the week of today.
export const gridLayout = (range: GridRange, today: Date = new Date()): GridLayout => {
  const months = GRID_RANGES.find((r) => r.value === range)?.months ?? 1
  const rangeStart = startOfDay(today)
  rangeStart.setMonth(rangeStart.getMonth() - months)
  const daysSinceMonday = (rangeStart.getDay() + 6) % 7
  const from = addDays(rangeStart, -daysSinceMonday)
  const to = addDays(startOfDay(today), 1)

  const weeks: string[][] = []
  for (let monday = from; monday < to; monday = addDays(monday, 7)) {
    weeks.push(Array.from({ length: 7 }, (_, i) => toDateString(addDays(monday, i))))
  }
  return { from: startOfDay(from), to, firstDay: toDateString(from), lastDay: toDateString(today), weeks }
}

// Groups entries by local day, then by user and activity.
export const groupByDay = (entries: Entry[]): Map<string, Day> => {
  const days = new Map<string, Day>()
  for (const entry of entries) {
    const date = toDateString(new Date(entry.occurredAt))
    let day = days.get(date)
    if (!day) {
      day = { date, items: [], entries: [], done: 0, excused: 0 }
      days.set(date, day)
    }

    day.entries.push(entry)

    let item = day.items.find((i) => i.userId === entry.userId && i.activityId === entry.activityId)
    if (!item) {
      item = { userId: entry.userId, activityId: entry.activityId, done: 0, excused: 0 }
      day.items.push(item)
    }

    if (entry.excused) {
      item.excused++
      day.excused++
    } else {
      item.done++
      day.done++
    }
  }
  for (const day of days.values()) {
    day.entries.sort((a, b) => a.occurredAt.localeCompare(b.occurredAt))
  }
  return days
}

// Intensity 0-4 from the number of completed entries, like GitHub.
export const level = (done: number): number => Math.min(done, 4)

export const scaledLevel = (done: number, max: number): number =>
  max <= 4 ? level(done) : Math.ceil((4 * done) / max)

export interface DayPerson {
  userId: string
  done: number
  excusedOnly: boolean
}

// People active on the day, most completed entries first.
export const peopleOnDay = (day: Day): DayPerson[] => {
  const byUser = new Map<string, DayPerson>()
  for (const item of day.items) {
    const person = byUser.get(item.userId) ?? { userId: item.userId, done: 0, excusedOnly: true }
    person.done += item.done
    person.excusedOnly = person.done === 0
    byUser.set(item.userId, person)
  }
  return [...byUser.values()].sort((a, b) => b.done - a.done)
}
