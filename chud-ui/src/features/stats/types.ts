export type OccurrenceStatus = 'done' | 'excused' | 'missed' | 'pending'

export interface Occurrence {
  date: string
  planId: string
  planTitle: string
  userId: string
  activityId: string
  status: OccurrenceStatus
  entryId: string | null
}

export interface UserStats {
  userId: string
  points: number
  done: number
  extra: number
  excused: number
  missed: number
  pending: number
  rate: number | null
}
