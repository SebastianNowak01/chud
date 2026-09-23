export interface Plan {
  id: string
  activityId: string
  userId: string
  title: string
  monday: boolean
  tuesday: boolean
  wednesday: boolean
  thursday: boolean
  friday: boolean
  saturday: boolean
  sunday: boolean
  startsOn: string
  endsOn: string | null
  createdAt: string
}

export type Weekday = 'monday' | 'tuesday' | 'wednesday' | 'thursday' | 'friday' | 'saturday' | 'sunday'

export const WEEKDAYS: { key: Weekday; label: string }[] = [
  { key: 'monday', label: 'Pn' },
  { key: 'tuesday', label: 'Wt' },
  { key: 'wednesday', label: 'Śr' },
  { key: 'thursday', label: 'Cz' },
  { key: 'friday', label: 'Pt' },
  { key: 'saturday', label: 'Sb' },
  { key: 'sunday', label: 'Nd' },
]

export interface PlanUpdatePayload {
  title?: string
  endsOn?: string
}

export type PlanPayload = Pick<Plan, 'title' | Weekday> & {
  startsOn: string // YYYY-MM-DD
  endsOn: string | null
}
