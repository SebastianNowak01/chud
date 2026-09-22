import type { Entry } from '@/features/entries/types'
import { WEEKDAYS, type Plan } from '@/features/plans/types'
import { addDays, apiDate, parseDate, toDateString } from '@/lib/dates'

export type OccurrenceStatus = 'done' | 'excused' | 'missed' | 'todo'

export interface Occurrence {
  date: string // YYYY-MM-DD
  status: OccurrenceStatus
  entry?: Entry
}

// JS getDay() is 0 = Sunday; WEEKDAYS starts at Monday.
const weekdayKey = (date: Date) => WEEKDAYS[(date.getDay() + 6) % 7].key

// Planned days of the plan within [from, to], each matched with the entry that resolved it.
export const planOccurrences = (plan: Plan, entries: Entry[], from: Date, to: Date): Occurrence[] => {
  const today = toDateString(new Date())
  const startsOn = parseDate(apiDate(plan.startsOn))
  const endsOn = plan.endsOn ? parseDate(apiDate(plan.endsOn)) : null

  const first = startsOn > from ? startsOn : from
  const last = endsOn && endsOn < to ? endsOn : to

  const occurrences: Occurrence[] = []
  for (let day = first; toDateString(day) <= toDateString(last); day = addDays(day, 1)) {
    if (!plan[weekdayKey(day)]) {
      continue
    }
    const date = toDateString(day)
    const entry = entries.find((e) => e.planId === plan.id && e.scheduledFor && apiDate(e.scheduledFor) === date)

    let status: OccurrenceStatus
    if (entry) {
      status = entry.excused ? 'excused' : 'done'
    } else {
      status = date < today ? 'missed' : 'todo'
    }
    occurrences.push({ date, status, entry })
  }
  return occurrences
}
