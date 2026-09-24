import { describe, expect, it } from 'vitest'
import { weekFace, weekHeading, weeklyCounts, weekScore } from '@/features/stats/mood'
import type { Entry } from '@/features/entries/types'
import type { Occurrence } from '@/features/stats/types'
import { currentWeek } from '@/lib/dates'

const counts = (done = 0, excused = 0, missed = 0, pending = 0, extra = 0) => ({ done, extra, excused, missed, pending })

describe('weekScore', () => {
  it.each([
    ['nothing planned', counts(), 0],
    ['week just started', counts(0, 0, 0, 5), 0],
    ['half done, half to go', counts(2, 0, 0, 2), 0.5],
    ['everything done', counts(5), 1],
    ['done and excuses cancel out', counts(1, 1), 0],
    ['all excuses', counts(0, 5), -1],
    ['misses weigh more than excuses', counts(1, 0, 1), -0.25],
    ['everything missed is clamped', counts(0, 0, 4), -1],
    ['extras add a little', counts(1, 1, 0, 0, 2), 0.1],
    ['extras alone', counts(0, 0, 0, 0, 3), 0.15],
    ['extras bonus is capped', counts(0, 0, 0, 0, 20), 0.25],
    ['extras cannot go past the top', counts(5, 0, 0, 0, 5), 1],
  ] as const)('%s', (_, week, score) => {
    expect(weekScore(week)).toBeCloseTo(score)
  })
})

describe('weekFace', () => {
  it.each([
    [-1, 0],
    [-0.5, 2],
    [0, 4],
    [0.3, 5],
    [0.5, 6],
    [1, 8],
  ])('%s shows face %s', (score, face) => {
    expect(weekFace(score)).toBe(face)
  })
})

describe('currentWeek', () => {
  it.each([
    [new Date(2026, 8, 21, 9), '2026-09-21'],
    [new Date(2026, 8, 24, 15), '2026-09-21'],
    [new Date(2026, 8, 27, 23), '2026-09-21'],
  ])('%s is in the week of %s', (today, monday) => {
    expect(currentWeek(today)).toEqual({ from: monday, to: '2026-09-27' })
  })
})

const occurrence = (date: string, status: Occurrence['status'], userId = 'alice'): Occurrence => ({
  date,
  planId: 'plan',
  planTitle: 'Siłownia',
  userId,
  activityId: 'gym',
  status,
  entryId: null,
})

const entry = (overrides: Partial<Entry>): Entry => ({
  id: crypto.randomUUID(),
  activityId: 'gym',
  userId: 'alice',
  planId: null,
  scheduledFor: null,
  excused: false,
  description: '',
  occurredAt: '',
  createdAt: '',
  ...overrides,
})

describe('weeklyCounts', () => {
  it('groups one user days by the Monday of their week', () => {
    const weeks = weeklyCounts(
      [
        occurrence('2026-09-27', 'done'),
        occurrence('2026-09-28', 'excused'),
        occurrence('2026-09-30', 'missed'),
        occurrence('2026-10-04', 'pending'),
        occurrence('2026-09-29', 'done', 'bob'),
      ],
      [
        entry({ occurredAt: new Date(2026, 8, 22, 18).toISOString() }),
        entry({ occurredAt: new Date(2026, 8, 29, 18).toISOString() }),
        entry({ occurredAt: new Date(2026, 8, 29, 19).toISOString(), planId: 'plan', scheduledFor: '2026-09-29' }),
        entry({ occurredAt: new Date(2026, 8, 30, 18).toISOString(), userId: 'bob' }),
      ],
      'alice',
    )
    expect([...weeks.entries()]).toEqual([
      ['2026-09-21', { done: 1, extra: 1, excused: 0, missed: 0, pending: 0 }],
      ['2026-09-28', { done: 0, extra: 1, excused: 1, missed: 1, pending: 1 }],
    ])
  })
})

describe('weekHeading', () => {
  it.each([
    ['2026-09-21', 'Status tygodnia'],
    ['2026-09-14', 'Status z zeszłego tygodnia'],
    ['2026-09-07', 'Status tygodnia 7–13 wrz'],
    ['2026-08-31', 'Status tygodnia 31 sie – 6 wrz'],
  ])('%s', (week, heading) => {
    expect(weekHeading(week, '2026-09-21')).toBe(heading)
  })
})
