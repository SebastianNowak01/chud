export interface Activity {
  id: string
  name: string
  description: string
  createdBy: string
  createdAt: string
  archivedAt: string | null
  hasHistory: boolean
}
