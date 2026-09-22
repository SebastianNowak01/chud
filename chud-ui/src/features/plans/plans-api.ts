import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { authFetch } from '@/lib/api-client'
import { activitiesKey } from '@/features/activities/activities-api'
import type { Plan, PlanPayload } from '@/features/plans/types'

const getPlans = async (activityId: string): Promise<Plan[]> => {
  const res = await authFetch(`/activities/${encodeURIComponent(activityId)}/plans`)
  return (await res.json()) as Plan[]
}

const createPlan = async (activityId: string, payload: PlanPayload): Promise<Plan> => {
  const res = await authFetch(`/activities/${encodeURIComponent(activityId)}/plans`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return (await res.json()) as Plan
}

const deletePlan = async (planId: string): Promise<void> => {
  await authFetch(`/plans/${encodeURIComponent(planId)}`, { method: 'DELETE' })
}

const plansKey = (activityId: string) => [...activitiesKey, activityId, 'plans']

export const usePlans = (activityId: string) =>
  useQuery({ queryKey: plansKey(activityId), queryFn: () => getPlans(activityId) })

export const useCreatePlan = (activityId: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: PlanPayload) => createPlan(activityId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: plansKey(activityId) }),
  })
}

export const useDeletePlan = (activityId: string) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deletePlan,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: plansKey(activityId) }),
  })
}
