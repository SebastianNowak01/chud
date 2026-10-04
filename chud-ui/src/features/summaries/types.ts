export type SummaryStatus = 'ready' | 'generating' | 'unavailable' | 'disabled' | 'empty'

export interface WeekSummary {
  week: string
  status: SummaryStatus
  text: string
  generatedAt: string | null
  outdated: boolean | null
  regenerateAvailableAt: string | null
}
