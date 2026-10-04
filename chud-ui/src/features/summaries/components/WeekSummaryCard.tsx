import { useEffect, useState } from 'react'
import { DrawablyButton } from 'drawably/react'
import { Card } from '@/components/ui/Card'
import { Hint } from '@/components/ui/Hint'
import { LoadError } from '@/components/ui/LoadError'
import { useRegenerateSummary, useWeekSummary } from '@/features/summaries/summaries-api'
import { canRegenerate, regenerateWaitMs, shiftWeek, weekLabel } from '@/features/summaries/summary'
import type { WeekSummary } from '@/features/summaries/types'
import { buttonState } from '@/lib/button-state'
import { currentWeek, formatDateTime } from '@/lib/dates'

const NAV_BUTTON =
  'grid size-9 cursor-pointer place-items-center rounded-full border-0 bg-transparent text-[18px] hover:bg-ink/8 disabled:cursor-default disabled:opacity-30 disabled:hover:bg-transparent'

const COUNTDOWN_TICK_MS = 15_000

const useCountdownNow = (availableAt: string | null | undefined) => {
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    if (!availableAt) return
    const tick = () => setNow(Date.now())
    const first = setTimeout(tick, 0)
    const interval = setInterval(tick, COUNTDOWN_TICK_MS)
    const end = setTimeout(
      () => {
        tick()
        clearInterval(interval)
      },
      Math.max(0, new Date(availableAt).getTime() - Date.now()) + 250,
    )
    return () => {
      clearTimeout(first)
      clearInterval(interval)
      clearTimeout(end)
    }
  }, [availableAt])
  return now
}

function SummaryBody({ summary }: { summary: WeekSummary }) {
  const generating = summary.status === 'generating'
  return (
    <>
      {summary.text && (
        <p className={`m-0 text-[16px] leading-relaxed ${generating ? 'opacity-50' : ''}`}>{summary.text}</p>
      )}
      {generating && <Hint className="m-0 animate-pulse">Model myśli… to może potrwać kilka minut.</Hint>}
      {summary.status === 'empty' && <Hint className="m-0">W tym tygodniu nie ma czego podsumować.</Hint>}
      {summary.status === 'unavailable' && (
        <Hint className="m-0">
          {summary.text ? 'Nie udało się odświeżyć podsumowania.' : 'Podsumowanie jest chwilowo niedostępne.'}
        </Hint>
      )}
      {summary.generatedAt && !generating && (
        <Hint className="m-0">
          Wygenerowano {formatDateTime(summary.generatedAt)}
          {summary.outdated && ' · dane od tego czasu się zmieniły'}
        </Hint>
      )}
    </>
  )
}

export function WeekSummaryCard() {
  const thisWeek = currentWeek().from
  const [week, setWeek] = useState(thisWeek)
  const summary = useWeekSummary(week)
  const regenerate = useRegenerateSummary()
  const now = useCountdownNow(summary.data?.regenerateAvailableAt)

  if (summary.data?.status === 'disabled') return null

  const waitMinutes = summary.data ? Math.ceil(regenerateWaitMs(summary.data, now) / 60_000) : 0

  return (
    <Card size="lg" className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center gap-2">
        <h3 className="flex-1">Podsumowanie tygodnia</h3>
        <div className="flex items-center gap-1">
          <button
            type="button"
            className={NAV_BUTTON}
            aria-label="Poprzedni tydzień"
            onClick={() => setWeek(shiftWeek(week, -1))}
          >
            ‹
          </button>
          <span className="min-w-28 text-center text-[14px] text-muted">{weekLabel(week, thisWeek)}</span>
          <button
            type="button"
            className={NAV_BUTTON}
            aria-label="Następny tydzień"
            disabled={week === thisWeek}
            onClick={() => setWeek(shiftWeek(week, 1))}
          >
            ›
          </button>
        </div>
      </div>

      {summary.isPending && <Hint className="m-0">Ładowanie…</Hint>}
      <LoadError error={summary.error} onRetry={() => void summary.refetch()} />
      {summary.data && (
        <div aria-live="polite" className="flex flex-col gap-2">
          <SummaryBody summary={summary.data} />
        </div>
      )}

      {summary.data && summary.data.status !== 'empty' && (
        <div className="flex flex-wrap items-center gap-3">
          <DrawablyButton
            tone="neutral"
            state={buttonState(regenerate.status)}
            disabled={!canRegenerate(summary.data, now) || regenerate.isPending}
            onClick={() => regenerate.mutate(week)}
          >
            {summary.data.text ? 'Wygeneruj ponownie' : 'Wygeneruj'}
          </DrawablyButton>
          {waitMinutes > 0 && summary.data.status !== 'generating' && (
            <Hint as="span">Ponownie za {waitMinutes} min</Hint>
          )}
        </div>
      )}
    </Card>
  )
}
