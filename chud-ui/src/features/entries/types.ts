export interface Entry {
  id: string
  activityId: string
  userId: string
  planId: string | null
  scheduledFor: string | null
  excused: boolean
  description: string
  occurredAt: string
  createdAt: string
}

export interface Media {
  id: string
  entryId: string
  contentType: string
}

export interface EntryPayload {
  description: string
  occurredAt: string // RFC 3339
  planId?: string
  scheduledFor?: string // YYYY-MM-DD
  excused?: boolean
  files: File[]
}
