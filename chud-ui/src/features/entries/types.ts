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

export const MEDIA_TYPES = [
  'image/jpeg',
  'image/png',
  'image/gif',
  'image/webp',
  'video/mp4',
  'video/webm',
  'video/quicktime',
]
export const MAX_FILE_SIZE = 10 * 1024 * 1024
export const MAX_FILE_COUNT = 10
export const MAX_DESCRIPTION_LENGTH = 2000
