import { useQuery } from '@tanstack/react-query'
import { authFetch } from '@/lib/api-client'
import type { Occurrence, UserStats } from '@/features/stats/types'

export const statsKey = ['stats']

const dateRange = (from: string, to: string) => `from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`

const getOccurrences = async (from: string, to: string): Promise<Occurrence[]> => {
  const res = await authFetch(`/occurrences?${dateRange(from, to)}`)
  return (await res.json()) as Occurrence[]
}

const getLeaderboard = async (from: string, to: string): Promise<UserStats[]> => {
  const res = await authFetch(`/stats?${dateRange(from, to)}`)
  return (await res.json()) as UserStats[]
}

export const useOccurrences = (from: string, to: string) =>
  useQuery({ queryKey: [...statsKey, 'occurrences', from, to], queryFn: () => getOccurrences(from, to) })

export const useLeaderboard = (from: string, to: string) =>
  useQuery({ queryKey: [...statsKey, 'leaderboard', from, to], queryFn: () => getLeaderboard(from, to) })
