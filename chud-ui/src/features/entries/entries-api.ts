import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { API_URL, authFetch } from '@/lib/api-client'
import { activitiesKey } from '@/features/activities/activities-api'
import type { Entry, EntryPayload, Media } from '@/features/entries/types'
import { statsKey } from '@/features/stats/stats-api'

const getEntries = async (activityId: string): Promise<Entry[]> => {
  const res = await authFetch(`/activities/${encodeURIComponent(activityId)}/entries`)
  return (await res.json()) as Entry[]
}

const createEntry = async (activityId: string, payload: EntryPayload): Promise<Entry> => {
  const form = new FormData()
  form.set('description', payload.description)
  form.set('occurredAt', payload.occurredAt)
  if (payload.planId) form.set('planId', payload.planId)
  if (payload.scheduledFor) form.set('scheduledFor', payload.scheduledFor)
  if (payload.excused) form.set('excused', 'true')
  for (const file of payload.files) form.append('files', file)

  const res = await authFetch(`/activities/${encodeURIComponent(activityId)}/entries`, {
    method: 'POST',
    body: form,
  })
  return (await res.json()) as Entry
}

const getEntryMedia = async (entryId: string): Promise<Media[]> => {
  const res = await authFetch(`/entries/${encodeURIComponent(entryId)}/media`)
  return (await res.json()) as Media[]
}

const rangeQuery = (from: string, to: string) => `from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`

const getEntriesInRange = async (from: string, to: string): Promise<Entry[]> => {
  const res = await authFetch(`/entries?${rangeQuery(from, to)}`)
  return (await res.json()) as Entry[]
}

const getMyEntriesInRange = async (from: string, to: string): Promise<Entry[]> => {
  const res = await authFetch(`/me/entries?${rangeQuery(from, to)}`)
  return (await res.json()) as Entry[]
}

const getUserEntriesInRange = async (userId: string, from: string, to: string): Promise<Entry[]> => {
  const res = await authFetch(`/users/${encodeURIComponent(userId)}/entries?${rangeQuery(from, to)}`)
  return (await res.json()) as Entry[]
}

// Media files are authenticated by the login cookie, so plain <img>/<video> tags can load them.
export const mediaUrl = (mediaId: string) => `${API_URL}/media/${encodeURIComponent(mediaId)}`

export const useEntries = (activityId: string) =>
  useQuery({ queryKey: [...activitiesKey, activityId, 'entries'], queryFn: () => getEntries(activityId) })

export const useEntryMedia = (entryId: string) =>
  useQuery({ queryKey: ['entries', entryId, 'media'], queryFn: () => getEntryMedia(entryId) })

export const useEntriesInRange = (from: string, to: string) =>
  useQuery({
    queryKey: ['entries', 'range', from, to],
    queryFn: () => getEntriesInRange(from, to),
  })

export const useUserEntriesInRange = (userId: string, from: string, to: string) =>
  useQuery({
    queryKey: ['entries', 'range', 'user', userId, from, to],
    queryFn: () => getUserEntriesInRange(userId, from, to),
  })

export const useMyEntriesInRange = (from: string, to: string) =>
  useQuery({
    queryKey: ['me', 'entries', from, to],
    queryFn: () => getMyEntriesInRange(from, to),
  })

export const useCreateEntry = (activityId: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: EntryPayload) => createEntry(activityId, payload),
    meta: { success: 'Zapisano wpis.', inlineError: true },
    // Entries, members and plan progress of the activity change, and so do the activity grids.
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: [...activitiesKey, activityId] }),
        queryClient.invalidateQueries({ queryKey: ['entries', 'range'] }),
        queryClient.invalidateQueries({ queryKey: ['me', 'entries'] }),
        queryClient.invalidateQueries({ queryKey: statsKey }),
      ]),
  })
}
