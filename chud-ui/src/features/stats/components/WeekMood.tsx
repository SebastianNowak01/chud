import { LoadError } from '@/components/ui/LoadError'
import { MoodAxis } from '@/features/stats/components/MoodAxis'
import { EMPTY_WEEK, weekHeading, weekScore, type WeekCounts } from '@/features/stats/mood'
import { currentWeek } from '@/lib/dates'

interface WeekMoodProps {
  week: string
  counts: WeekCounts | undefined
  loaded: boolean
  error?: Error | null
  onRetry?: () => void
}

export function WeekMood({ week, counts = EMPTY_WEEK, loaded, error, onRetry }: WeekMoodProps) {
  return (
    <div className="flex w-full flex-col gap-2">
      <span className="text-[14px] font-semibold tracking-widest text-muted uppercase">
        {weekHeading(week, currentWeek().from)}
      </span>
      <MoodAxis score={weekScore(counts)} />
      <LoadError error={error} onRetry={onRetry} />
      {loaded && (
        <span className="text-[13px] text-muted">
          zrobione {counts.done} · poza planem {counts.extra} · wymówki {counts.excused} · opuszczone {counts.missed} · do
          zrobienia {counts.pending}
        </span>
      )}
    </div>
  )
}
