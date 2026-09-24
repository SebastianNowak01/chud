import { describe, expect, it } from 'vitest'
import { filterEntries, parseIds, serializeIds, toggleId } from '@/features/dashboard/filters'
import type { Entry } from '@/features/entries/types'

const entry = (userId: string, activityId: string): Entry => ({
  id: `${userId}-${activityId}`,
  activityId,
  userId,
  planId: null,
  scheduledFor: null,
  excused: false,
  description: '',
  occurredAt: '',
  createdAt: '',
})

const entries = [entry('ala', 'gym'), entry('ala', 'run'), entry('bob', 'gym'), entry('cyd', 'read')]
const ids = (list: Entry[]) => list.map((e) => e.id)

describe('filterEntries', () => {
  it('keeps everything without a filter', () => {
    expect(filterEntries(entries, { userIds: [], activityIds: [] })).toEqual(entries)
  })

  it('keeps any of the chosen people', () => {
    expect(ids(filterEntries(entries, { userIds: ['ala', 'cyd'], activityIds: [] }))).toEqual([
      'ala-gym',
      'ala-run',
      'cyd-read',
    ])
  })

  it('keeps any of the chosen activities', () => {
    expect(ids(filterEntries(entries, { userIds: [], activityIds: ['gym'] }))).toEqual(['ala-gym', 'bob-gym'])
  })

  it('needs both a chosen person and a chosen activity when both are set', () => {
    expect(ids(filterEntries(entries, { userIds: ['ala'], activityIds: ['gym', 'read'] }))).toEqual(['ala-gym'])
  })
})

describe('ids in the address', () => {
  it('parses a comma list, dropping blanks and repeats', () => {
    expect(parseIds('a,,b,a')).toEqual(['a', 'b'])
  })

  it('ignores anything that is not a string', () => {
    expect(parseIds(undefined)).toEqual([])
    expect(parseIds(['a'])).toEqual([])
  })

  it('leaves an empty filter out of the address', () => {
    expect(serializeIds([])).toBeUndefined()
    expect(serializeIds(['a', 'b'])).toBe('a,b')
  })

  it('toggles an id', () => {
    expect(toggleId(['a'], 'b')).toEqual(['a', 'b'])
    expect(toggleId(['a', 'b'], 'a')).toEqual(['b'])
  })
})
