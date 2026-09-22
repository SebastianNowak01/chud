import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { authFetch } from '@/lib/api-client'
import type { Activity } from '@/features/activities/types'
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
    onSuccess: () => queryClient.invalidateQueries({ queryKey: activitiesKey }),
  })
}
