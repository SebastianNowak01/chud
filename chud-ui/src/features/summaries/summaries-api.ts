import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { authFetch } from '@/lib/api-client'
import type { WeekSummary } from '@/features/summaries/types'

export const summariesKey = ['summaries'] as const

export const POLL_INTERVAL_MS = 3000

const weekQuery = (week: string) => `week=${encodeURIComponent(week)}`

const getWeekSummary = async (week: string): Promise<WeekSummary> => {
  const res = await authFetch(`/summaries/week?${weekQuery(week)}`)
  return (await res.json()) as WeekSummary
}

const regenerateWeekSummary = async (week: string): Promise<WeekSummary> => {
  const res = await authFetch(`/summaries/week/regenerate?${weekQuery(week)}`, {
    method: 'POST',
  })
  return (await res.json()) as WeekSummary
}

export const useWeekSummary = (week: string) =>
  useQuery({
    queryKey: [...summariesKey, week],
    queryFn: () => getWeekSummary(week),
    refetchInterval: (query) => (query.state.data?.status === 'generating' ? POLL_INTERVAL_MS : false),
  })

export const useRegenerateSummary = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: regenerateWeekSummary,
    onSuccess: (summary) => queryClient.setQueryData([...summariesKey, summary.week], summary),
  })
}
