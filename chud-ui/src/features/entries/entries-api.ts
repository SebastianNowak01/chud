import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { API_URL, authFetch } from '@/lib/api-client'
import { activitiesKey } from '@/features/activities/activities-api'
import type { Entry, EntryPayload, Media } from '@/features/entries/types'

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

const rangeQuery = (from: Date, to: Date) =>
  `from=${encodeURIComponent(from.toISOString())}&to=${encodeURIComponent(to.toISOString())}`

const getEntriesInRange = async (from: Date, to: Date): Promise<Entry[]> => {
  const res = await authFetch(`/entries?${rangeQuery(from, to)}`)
  return (await res.json()) as Entry[]
}

const getMyEntriesInRange = async (from: Date, to: Date): Promise<Entry[]> => {
  const res = await authFetch(`/me/entries?${rangeQuery(from, to)}`)
  return (await res.json()) as Entry[]
}

// Media files are authenticated by the login cookie, so plain <img>/<video> tags can load them.
export const mediaUrl = (mediaId: string) => `${API_URL}/media/${encodeURIComponent(mediaId)}`

export const useEntries = (activityId: string) =>
  useQuery({ queryKey: [...activitiesKey, activityId, 'entries'], queryFn: () => getEntries(activityId) })

export const useEntryMedia = (entryId: string) =>
  useQuery({ queryKey: ['entries', entryId, 'media'], queryFn: () => getEntryMedia(entryId) })

// Everyone's entries in [from, to).
export const useEntriesInRange = (from: Date, to: Date) =>
  useQuery({
    queryKey: ['entries', 'range', from.toISOString(), to.toISOString()],
    queryFn: () => getEntriesInRange(from, to),
  })

// The current user's entries in [from, to).
export const useMyEntriesInRange = (from: Date, to: Date) =>
  useQuery({
    queryKey: ['me', 'entries', from.toISOString(), to.toISOString()],
    queryFn: () => getMyEntriesInRange(from, to),
  })

export const useCreateEntry = (activityId: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: EntryPayload) => createEntry(activityId, payload),
    // Entries, members and plan progress of the activity change, and so do the activity grids.
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: [...activitiesKey, activityId] }),
        queryClient.invalidateQueries({ queryKey: ['entries', 'range'] }),
        queryClient.invalidateQueries({ queryKey: ['me', 'entries'] }),
      ]),
  })
}
