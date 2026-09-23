import { useMemo, useState, type KeyboardEvent } from 'react'
import { DrawablySelect } from 'drawably/react'
import { roughRoundedRect, scribbleFill } from 'drawably'
import { UserTag } from '@/components/common/UserTag'
import { Card, type CardSize } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { Hint } from '@/components/ui/Hint'
import type { Activity } from '@/features/activities/types'
import type { Entry } from '@/features/entries/types'
import type { User } from '@/features/users/types'
import {
  GRID_RANGES,
  groupByDay,
  level,
  peopleOnDay,
  type Day,
  type GridLayout,
  type GridRange,
} from '@/lib/contributions'
import { formatDate, formatMonth, toDateString } from '@/lib/dates'

// The grid is one SVG drawn in "units"; each range renders a unit at a fixed pixel size,
// and on narrow screens the whole SVG scales down instead of scrolling.
const CELL = 20
const STEP = 24 // cell + gap
const LABEL_WIDTH = 30
const HEADER_HEIGHT = 18
const PX_PER_UNIT: Record<GridRange, number> = { month: 1.5, quarter: 1.2, year: 0.68 }

// Pen stroke per intensity level 1-4: heavier and darker the more was done.
const LEVEL_STROKE = [0, 1.6, 2.4, 3.2, 4.2]
const LEVEL_OPACITY = [0, 0.45, 0.65, 0.85, 1]
const DAY_LABELS = ['Mon', '', 'Wed', '', 'Fri', '', '']

// 'single' colors every day with one color; 'people' splits a day into stripes of each active person's color.
export type GridColoring = { kind: 'single'; color: string } | { kind: 'people' }

// Sketches are deterministic per day, so each is generated once and reused across renders.
const sketches = new Map<string, string>()
const sketch = (key: string, draw: () => string) => {
  let path = sketches.get(key)
  if (path === undefined) {
    path = draw()
    sketches.set(key, path)
  }
  return path
}

const seedOf = (text: string) => {
  let hash = 2166136261
  for (let i = 0; i < text.length; i++) {
    hash = Math.imul(hash ^ text.charCodeAt(i), 16777619)
  }
  return hash >>> 0
}

interface Stripe {
  color: string
  done: number // 0 = only excused
}

const INK = 'fill-none [stroke-linecap:round] [stroke-linejoin:round]'
const LABEL = 'fill-muted text-[11px]'

interface OutlineState {
  future?: boolean
  today?: boolean
  selected?: boolean
  faded?: boolean
}

const outlineClass = ({ future, today, selected, faded }: OutlineState) => {
  const stroke = selected
    ? 'stroke-ink [stroke-width:2.4]'
    : today
      ? 'stroke-muted [stroke-width:1.6]'
      : future
        ? 'stroke-grid-faint [stroke-width:1]'
        : 'stroke-grid [stroke-width:1]'
  return [
    'fill-none',
    stroke,
    future && '[stroke-dasharray:2_3]',
    faded && '[stroke-opacity:0.55]',
  ]
    .filter(Boolean)
    .join(' ')
}

// One hand-drawn cell in local coordinates (0..CELL): pen scribbles for each stripe, then a rough outline.
function SketchCell({ seedKey, stripes, outline: state = {} }: { seedKey: string; stripes: Stripe[]; outline?: OutlineState }) {
  const seed = seedOf(seedKey)
  const outline = sketch(`outline:${seed}`, () => roughRoundedRect(1, 1, CELL - 2, CELL - 2, 4, { seed, roughness: 0.7 }))
  const width = (CELL - 4) / Math.max(stripes.length, 1)

  return (
    <>
      {stripes.map((stripe, i) => {
        const fill = sketch(`fill:${seed}:${i}:${stripes.length}`, () =>
          scribbleFill(2 + i * width, 2, width, CELL - 4, { seed: seed + i, roughness: 0.8 }),
        )
        const lvl = level(stripe.done)
        return stripe.done > 0 ? (
          <path
            key={i}
            d={fill}
            stroke={stripe.color}
            strokeWidth={LEVEL_STROKE[lvl]}
            strokeOpacity={LEVEL_OPACITY[lvl]}
            className={INK}
          />
        ) : (
          <path key={i} d={fill} stroke={stripe.color} className={`${INK} [stroke-dasharray:2_2.5] [stroke-opacity:0.8] [stroke-width:1.2]`} />
        )
      })}
      <path d={outline} className={outlineClass(state)} />
    </>
  )
}

interface ContributionGridProps {
  title: string
  layout: GridLayout
  range: GridRange
  onRangeChange: (range: GridRange) => void
  entries: Entry[] | undefined
  error?: Error | null
  coloring: GridColoring
  usersById: Map<string, User>
  activitiesById: Map<string, Activity>
  size?: CardSize
}

export function ContributionGrid({
  title,
  layout,
  range,
  onRangeChange,
  entries,
  error,
  coloring,
  usersById,
  activitiesById,
  size,
}: ContributionGridProps) {
  const today = toDateString(new Date())
  const [selected, setSelected] = useState(today)
  const days = useMemo(() => groupByDay(entries ?? []), [entries])

  const width = LABEL_WIDTH + layout.weeks.length * STEP
  const height = HEADER_HEIGHT + 7 * STEP

  const monthLabels = layout.weeks.map((week, i) => {
    const labelDate = i === 0 ? week[0] : week.find((date) => date.endsWith('-01'))
    return labelDate ? formatMonth(labelDate) : ''
  })

  const stripesOf = (day: Day | undefined): Stripe[] => {
    if (!day?.items.length) {
      return []
    }
    if (coloring.kind === 'single') {
      return [{ color: coloring.color, done: day.done }]
    }
    // In the small year view stripes would be unreadable, so only the day's leader is drawn.
    const people = range === 'year' ? peopleOnDay(day).slice(0, 1) : peopleOnDay(day)
    return people.map((p) => ({
      color: usersById.get(p.userId)?.color ?? '#888888',
      done: p.excusedOnly ? 0 : day.done,
    }))
  }

  const select = (date: string) => (e?: KeyboardEvent) => {
    if (e && e.key !== 'Enter' && e.key !== ' ') return
    e?.preventDefault()
    setSelected(date)
  }

  return (
    <Card size={size} className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-4 [&_select]:min-h-11">
        <h3 className="flex-1">{title}</h3>
        <DrawablySelect
          aria-label="Range"
          value={range}
          onChange={(e) => onRangeChange(e.target.value as GridRange)}
        >
          {GRID_RANGES.map((r) => (
            <option key={r.value} value={r.value}>
              {r.label}
            </option>
          ))}
        </DrawablySelect>
      </div>

      {error && <ErrorText>{error.message}</ErrorText>}

      <svg
        className="block h-auto w-full"
        viewBox={`0 0 ${width} ${height}`}
        style={{ maxWidth: width * PX_PER_UNIT[range] }}
      >
        {monthLabels.map((label, i) => (
          <text key={i} x={LABEL_WIDTH + i * STEP} y={12} className={LABEL}>
            {label}
          </text>
        ))}
        {DAY_LABELS.map((label, weekday) => (
          <text key={weekday} x={0} y={HEADER_HEIGHT + weekday * STEP + 15} className={LABEL}>
            {label}
          </text>
        ))}

        {layout.weeks.map((week, col) =>
          week.map((date, weekday) => {
            const day = days.get(date)
            const isFuture = date > today
            const summary = day ? `${day.done} done${day.excused ? `, ${day.excused} excused` : ''}` : 'nothing'

            return (
              <g key={date} transform={`translate(${LABEL_WIDTH + col * STEP} ${HEADER_HEIGHT + weekday * STEP})`}>
                <SketchCell
                  seedKey={date}
                  stripes={stripesOf(day)}
                  outline={{
                    future: isFuture,
                    today: date === today,
                    selected: date === selected,
                    faded: range === 'year',
                  }}
                />
                {!isFuture && (
                  <rect
                    width={CELL}
                    height={CELL}
                    className="cursor-pointer fill-transparent outline-none focus-visible:stroke-ink focus-visible:[stroke-width:2]"
                    role="button"
                    tabIndex={0}
                    aria-label={`${formatDate(date)}: ${summary}`}
                    aria-pressed={date === selected}
                    onClick={() => select(date)()}
                    onKeyDown={select(date)}
                  >
                    <title>{`${formatDate(date)}: ${summary}`}</title>
                  </rect>
                )}
              </g>
            )
          }),
        )}
      </svg>

      <Legend />

      <DayDetails day={days.get(selected)} usersById={usersById} activitiesById={activitiesById} />
    </Card>
  )
}

function Legend() {
  const samples: Stripe[][] = [[], ...[1, 2, 3, 4].map((done) => [{ color: 'var(--color-ink)', done }])]
  const legendCell = (key: string, stripes: Stripe[], i: number) => (
    <g key={key} transform={`translate(${i * STEP} 0)`}>
      <SketchCell seedKey={key} stripes={stripes} />
    </g>
  )

  return (
    <div className="flex flex-wrap items-center gap-1 text-[11px] text-muted">
      <span>less</span>
      <svg viewBox={`0 0 ${5 * STEP - 4} ${CELL}`} className="h-[11px] w-auto">
        {samples.map((stripes, i) => legendCell(`legend-${i}`, stripes, i))}
      </svg>
      <span>more</span>
      <svg viewBox={`0 0 ${CELL} ${CELL}`} className="ml-2 size-[11px]">
        <SketchCell seedKey="legend-excused" stripes={[{ color: 'var(--color-ink)', done: 0 }]} />
      </svg>
      <span>excused</span>
    </div>
  )
}

interface DayDetailsProps {
  day: Day | undefined
  usersById: Map<string, User>
  activitiesById: Map<string, Activity>
}

function DayDetails({ day, usersById, activitiesById }: DayDetailsProps) {
  return (
    <div>
      {!day?.items.length ? (
        <Hint className="m-0">Nothing logged.</Hint>
      ) : (
        <ul className="m-0 list-none p-0">
          {day.items.map((item) => (
            <li
              key={`${item.userId}-${item.activityId}`}
              className="flex min-h-11 flex-wrap items-center gap-2 border-t border-dashed border-rule"
            >
              <UserTag user={usersById.get(item.userId)} />
              <span>{activitiesById.get(item.activityId)?.name ?? 'activity'}</span>
              {item.done > 1 && <Hint as="span">×{item.done}</Hint>}
              {item.excused > 0 && <Hint as="span">excused{item.excused > 1 ? ` ×${item.excused}` : ''}</Hint>}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
