import { describe, expect, it } from 'vitest'
import { canRegenerate, regenerateWaitMs, shiftWeek, weekLabel } from '@/features/summaries/summary'
import type { WeekSummary } from '@/features/summaries/types'

const summary = (overrides: Partial<WeekSummary> = {}): WeekSummary => ({
  week: '2026-09-28',
  status: 'ready',
  text: 'Tekst',
  generatedAt: '2026-09-30T10:00:00Z',
  outdated: false,
  regenerateAvailableAt: '2026-09-30T10:05:00Z',
  ...overrides,
})

const at = (iso: string) => new Date(iso).getTime()

describe('weekLabel', () => {
  it('names recent weeks', () => {
    expect(weekLabel('2026-09-28', '2026-09-28')).toBe('Ten tydzień')
    expect(weekLabel('2026-09-21', '2026-09-28')).toBe('Zeszły tydzień')
  })

  it('shows the range for older weeks', () => {
    expect(weekLabel('2026-09-07', '2026-09-28')).toBe('7–13 wrz')
    expect(weekLabel('2026-08-31', '2026-09-28')).toBe('31 sie – 6 wrz')
  })
})

describe('shiftWeek', () => {
  it('moves by whole weeks across months', () => {
    expect(shiftWeek('2026-09-28', 1)).toBe('2026-10-05')
    expect(shiftWeek('2026-09-28', -4)).toBe('2026-08-31')
  })
})

describe('regenerate availability', () => {
  it('waits for the cooldown', () => {
    expect(regenerateWaitMs(summary(), at('2026-09-30T10:04:00Z'))).toBe(60_000)
    expect(canRegenerate(summary(), at('2026-09-30T10:04:00Z'))).toBe(false)
    expect(canRegenerate(summary(), at('2026-09-30T10:05:00Z'))).toBe(true)
  })

  it('is available without a cooldown', () => {
    expect(canRegenerate(summary({ regenerateAvailableAt: null }), 0)).toBe(true)
  })

  it('is blocked while generating or when there is nothing to summarise', () => {
    for (const status of ['generating', 'disabled', 'empty'] as const) {
      expect(canRegenerate(summary({ status, regenerateAvailableAt: null }), 0)).toBe(false)
    }
  })
})
