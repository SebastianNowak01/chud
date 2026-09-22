import { describe, expect, it } from 'vitest'
import type { Entry } from '@/features/entries/types'
import { gridLayout, groupByDay, level, peopleOnDay } from '@/lib/contributions'
import { toDateString } from '@/lib/dates'

const entry = (overrides: Partial<Entry>): Entry => ({
  id: crypto.randomUUID(),
  activityId: 'gym',
  userId: 'alice',
  planId: null,
  scheduledFor: null,
  excused: false,
  description: '',
  occurredAt: new Date(2026, 8, 21, 18).toISOString(),
  createdAt: '',
  ...overrides,
})

describe('gridLayout', () => {
  const today = new Date(2026, 8, 23, 15) // Wed, 23 Sep 2026

  it.each([
    ['month', '2026-08-17'], // 23 Aug is a Sunday, its Monday is 17 Aug
    ['quarter', '2026-06-22'], // 23 Jun is a Tuesday
    ['year', '2025-09-22'], // 23 Sep 2025 is a Tuesday
  ] as const)('%s starts on the Monday on or before the range start', (range, firstDay) => {
    const layout = gridLayout(range, today)
    expect(layout.weeks[0][0]).toBe(firstDay)
    expect(toDateString(layout.from)).toBe(firstDay)
  })

  it('ends with the week of today and asks for entries up to tomorrow', () => {
    const layout = gridLayout('month', today)
    const lastWeek = layout.weeks.at(-1)!
    expect(lastWeek[0]).toBe('2026-09-21')
    expect(lastWeek).toContain('2026-09-23')
    expect(toDateString(layout.to)).toBe('2026-09-24')
  })

  it('builds whole weeks, Monday first', () => {
    for (const week of gridLayout('year', today).weeks) {
      expect(week).toHaveLength(7)
      const [y, m, d] = week[0].split('-').map(Number)
      expect(new Date(y, m - 1, d).getDay()).toBe(1)
    }
  })
})

describe('groupByDay', () => {
  it('groups by local day, user and activity, counting excuses separately', () => {
    const days = groupByDay([
      entry({}),
      entry({}),
      entry({ userId: 'bob', activityId: 'walk' }),
      entry({ userId: 'bob', activityId: 'laundry', excused: true }),
      entry({ occurredAt: new Date(2026, 8, 22, 9).toISOString() }),
    ])

    const day = days.get('2026-09-21')!
    expect(day.done).toBe(3)
    expect(day.excused).toBe(1)
    expect(day.items).toContainEqual({ userId: 'alice', activityId: 'gym', done: 2, excused: 0 })
    expect(day.items).toContainEqual({ userId: 'bob', activityId: 'laundry', done: 0, excused: 1 })
    expect(days.get('2026-09-22')!.done).toBe(1)
  })
})

describe('peopleOnDay', () => {
  it('orders people by completed entries and marks excuse-only ones', () => {
    const day = groupByDay([
      entry({ userId: 'bob', excused: true }),
      entry({ userId: 'carol' }),
      entry({ userId: 'alice' }),
      entry({ userId: 'alice', activityId: 'walk' }),
    ]).get('2026-09-21')!

    expect(peopleOnDay(day)).toEqual([
      { userId: 'alice', done: 2, excusedOnly: false },
      { userId: 'carol', done: 1, excusedOnly: false },
      { userId: 'bob', done: 0, excusedOnly: true },
    ])
  })
})

describe('level', () => {
  it('caps at 4', () => {
    expect([0, 1, 3, 4, 9].map(level)).toEqual([0, 1, 3, 4, 4])
  })
})
