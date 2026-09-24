import type { Entry } from '@/features/entries/types'

export interface DashboardFilter {
  userIds: string[]
  activityIds: string[]
}

export const parseIds = (value: unknown): string[] =>
  typeof value === 'string' ? [...new Set(value.split(',').filter(Boolean))] : []

export const serializeIds = (ids: string[]): string | undefined => (ids.length > 0 ? ids.join(',') : undefined)

export const toggleId = (ids: string[], id: string): string[] =>
  ids.includes(id) ? ids.filter((i) => i !== id) : [...ids, id]

export const filterEntries = (entries: Entry[], { userIds, activityIds }: DashboardFilter): Entry[] =>
  entries.filter(
    (entry) =>
      (userIds.length === 0 || userIds.includes(entry.userId)) &&
      (activityIds.length === 0 || activityIds.includes(entry.activityId)),
  )
