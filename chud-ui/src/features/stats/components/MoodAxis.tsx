import { useMemo } from 'react'
import { roughLine } from 'drawably'
import { MoodWojak } from '@/features/stats/components/MoodWojak'

const INK = 'fill-none stroke-ink [stroke-width:1.8] [stroke-linecap:round] [stroke-linejoin:round]'

export function MoodAxis({ score }: { score: number }) {
  const { line, tick, head } = useMemo(
    () => ({
      line: roughLine(0, 10, 100, 10, { seed: 3, roughness: 0.6 }),
      tick: roughLine(50, 3, 50, 17, { seed: 5, roughness: 0.4 }),
      head: `${roughLine(12, 2, 2, 10, { seed: 7, roughness: 0.4 })} ${roughLine(2, 10, 12, 18, { seed: 8, roughness: 0.4 })}`,
    }),
    [],
  )

  return (
    <div
      className="flex w-full flex-col gap-1"
      role="meter"
      aria-label="Od CHUDA do CHADA"
      aria-valuemin={-1}
      aria-valuemax={1}
      aria-valuenow={Math.round(score * 100) / 100}
    >
      <div className="relative h-[72px]">
        <div className="absolute inset-x-0 top-1/2 flex h-5 -translate-y-1/2" aria-hidden="true">
          <svg viewBox="0 0 14 20" className="h-5 w-3.5 flex-none overflow-visible">
            <path d={head} className={INK} />
          </svg>
          <svg viewBox="0 0 100 20" preserveAspectRatio="none" className="-mx-3.5 h-5 min-w-0 flex-1 overflow-visible">
            <path d={line} vectorEffect="non-scaling-stroke" className={INK} />
            <path d={tick} vectorEffect="non-scaling-stroke" className="fill-none stroke-muted [stroke-width:1.5]" />
          </svg>
          <svg viewBox="0 0 14 20" className="h-5 w-3.5 flex-none -scale-x-100 overflow-visible">
            <path d={head} className={INK} />
          </svg>
        </div>
        <div className="absolute inset-y-0 right-12 left-12">
          <MoodWojak
            score={score}
            className="absolute top-0 size-[72px] -translate-x-1/2 transition-[left] duration-300 ease-out"
            style={{ left: `${((score + 1) / 2) * 100}%` }}
          />
        </div>
      </div>
      <div className="flex justify-between text-[22px] font-semibold uppercase">
        <span>Chud</span>
        <span>Chad</span>
      </div>
    </div>
  )
}
