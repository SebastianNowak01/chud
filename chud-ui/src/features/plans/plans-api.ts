import { useMutation, type MutationMeta, useQuery, useQueryClient } from '@tanstack/react-query'
import { authFetch } from '@/lib/api-client'
import { activitiesKey } from '@/features/activities/activities-api'
import type { Plan, PlanPayload, PlanUpdatePayload } from '@/features/plans/types'
import { statsKey } from '@/features/stats/stats-api'

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

const updatePlan = async (planId: string, payload: PlanUpdatePayload): Promise<Plan> => {
  const res = await authFetch(`/plans/${encodeURIComponent(planId)}`, {
    method: 'PATCH',
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

const usePlanMutation = <T,>(
  activityId: string,
  mutationFn: (variables: T) => Promise<unknown>,
  meta: MutationMeta,
) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn,
    meta,
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: plansKey(activityId) }),
        queryClient.invalidateQueries({ queryKey: statsKey }),
      ]),
  })
}

export const useCreatePlan = (activityId: string) =>
  usePlanMutation(activityId, (payload: PlanPayload) => createPlan(activityId, payload), {
    success: 'Dodano plan.',
    inlineError: true,
  })

export const useUpdatePlan = (activityId: string) =>
  usePlanMutation(
    activityId,
    ({ planId, payload }: { planId: string; payload: PlanUpdatePayload }) => updatePlan(planId, payload),
    { success: 'Zakończono plan.' },
  )

export const useDeletePlan = (activityId: string) =>
  usePlanMutation(activityId, deletePlan, { success: 'Usunięto plan.' })
