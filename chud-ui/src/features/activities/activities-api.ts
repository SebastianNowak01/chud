import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { authFetch } from '@/lib/api-client'
import type { Activity } from '@/features/activities/types'
import { statsKey } from '@/features/stats/stats-api'
import type { User } from '@/features/users/types'

export const activitiesKey = ['activities'] as const

const getActivities = async (): Promise<Activity[]> => {
  const res = await authFetch('/activities')
  return (await res.json()) as Activity[]
}

const getActivity = async (id: string): Promise<Activity> => {
  const res = await authFetch(`/activities/${encodeURIComponent(id)}`)
  return (await res.json()) as Activity
}

const getMembers = async (id: string): Promise<User[]> => {
  const res = await authFetch(`/activities/${encodeURIComponent(id)}/members`)
  return (await res.json()) as User[]
}

const createActivity = async (payload: { name: string; description: string }): Promise<Activity> => {
  const res = await authFetch('/activities', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return (await res.json()) as Activity
}

export const useActivities = () => useQuery({ queryKey: activitiesKey, queryFn: getActivities })

export const useActivity = (id: string) =>
  useQuery({ queryKey: [...activitiesKey, id], queryFn: () => getActivity(id) })

export const useMembers = (id: string) =>
  useQuery({ queryKey: [...activitiesKey, id, 'members'], queryFn: () => getMembers(id) })

export const useCreateActivity = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: createActivity,
    meta: { success: 'Utworzono aktywność.', inlineError: true },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: activitiesKey }),
  })
}

const activityPath = (id: string) => `/activities/${encodeURIComponent(id)}`

const deleteActivity = async (id: string): Promise<void> => {
  await authFetch(activityPath(id), { method: 'DELETE' })
}

const archiveActivity = async (id: string): Promise<Activity> => {
  const res = await authFetch(`${activityPath(id)}/archive`, { method: 'POST' })
  return (await res.json()) as Activity
}

const restoreActivity = async (id: string): Promise<Activity> => {
  const res = await authFetch(`${activityPath(id)}/restore`, { method: 'POST' })
  return (await res.json()) as Activity
}

export const useDeleteActivity = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteActivity,
    meta: { success: 'Usunięto aktywność.', inlineError: true },
    onSuccess: (_data, id) => {
      queryClient.removeQueries({ queryKey: [...activitiesKey, id] })
      return queryClient.invalidateQueries({ queryKey: activitiesKey, exact: true })
    },
  })
}

const useArchiveMutation = (mutationFn: (id: string) => Promise<Activity>, success: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn,
    meta: { success, inlineError: true },
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: activitiesKey }),
        queryClient.invalidateQueries({ queryKey: statsKey }),
      ]),
  })
}

export const useArchiveActivity = () => useArchiveMutation(archiveActivity, 'Zarchiwizowano aktywność.')

export const useRestoreActivity = () => useArchiveMutation(restoreActivity, 'Przywrócono aktywność.')
