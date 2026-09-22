import { useMemo, useState, type KeyboardEvent } from 'react'
import { DrawablyCard, DrawablySelect } from 'drawably/react'
import { roughRoundedRect, scribbleFill } from 'drawably'
import { UserTag } from '@/components/common/UserTag'
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

// One hand-drawn cell in local coordinates (0..CELL): pen scribbles for each stripe, then a rough outline.
function SketchCell({ seedKey, stripes, className }: { seedKey: string; stripes: Stripe[]; className?: string }) {
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
            className="cgrid__ink"
          />
        ) : (
          <path key={i} d={fill} stroke={stripe.color} className="cgrid__ink cgrid__ink--excused" />
        )
      })}
      <path d={outline} className={`cgrid__outline ${className ?? ''}`} />
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
    <DrawablyCard className="card stack cgrid">
      <div className="header cgrid__header">
        <h3>{title}</h3>
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

      {error && <p className="error">{error.message}</p>}

      <svg
        className={`cgrid__svg cgrid__svg--${range}`}
        viewBox={`0 0 ${width} ${height}`}
        style={{ maxWidth: width * PX_PER_UNIT[range] }}
      >
        {monthLabels.map((label, i) => (
          <text key={i} x={LABEL_WIDTH + i * STEP} y={12} className="cgrid__label">
            {label}
          </text>
        ))}
        {DAY_LABELS.map((label, weekday) => (
          <text key={weekday} x={0} y={HEADER_HEIGHT + weekday * STEP + 15} className="cgrid__label">
            {label}
          </text>
        ))}

        {layout.weeks.map((week, col) =>
          week.map((date, weekday) => {
            const day = days.get(date)
            const isFuture = date > today
            const outlineClass = [
              isFuture && 'cgrid__outline--future',
              date === today && 'cgrid__outline--today',
              date === selected && 'cgrid__outline--selected',
            ]
              .filter(Boolean)
              .join(' ')
            const summary = day ? `${day.done} done${day.excused ? `, ${day.excused} excused` : ''}` : 'nothing'

            return (
              <g key={date} transform={`translate(${LABEL_WIDTH + col * STEP} ${HEADER_HEIGHT + weekday * STEP})`}>
                <SketchCell seedKey={date} stripes={stripesOf(day)} className={outlineClass} />
                {!isFuture && (
                  <rect
                    width={CELL}
                    height={CELL}
                    className="cgrid__hit"
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
    </DrawablyCard>
  )
}

function Legend() {
  const samples: Stripe[][] = [[], ...[1, 2, 3, 4].map((done) => [{ color: 'var(--text)', done }])]
  const legendCell = (key: string, stripes: Stripe[], i: number) => (
    <g key={key} transform={`translate(${i * STEP} 0)`}>
      <SketchCell seedKey={key} stripes={stripes} />
    </g>
  )

  return (
    <div className="cgrid__legend">
      <span>less</span>
      <svg viewBox={`0 0 ${5 * STEP - 4} ${CELL}`} className="cgrid__legend-cells">
        {samples.map((stripes, i) => legendCell(`legend-${i}`, stripes, i))}
      </svg>
      <span>more</span>
      <svg viewBox={`0 0 ${CELL} ${CELL}`} className="cgrid__legend-cell">
        <SketchCell seedKey="legend-excused" stripes={[{ color: 'var(--text)', done: 0 }]} />
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
    <div className="cgrid__details">
      {!day?.items.length ? (
        <p className="hint">Nothing logged.</p>
      ) : (
        <ul>
          {day.items.map((item) => (
            <li key={`${item.userId}-${item.activityId}`}>
              <UserTag user={usersById.get(item.userId)} />
              <span>{activitiesById.get(item.activityId)?.name ?? 'activity'}</span>
              {item.done > 1 && <span className="hint">×{item.done}</span>}
              {item.excused > 0 && <span className="hint">excused{item.excused > 1 ? ` ×${item.excused}` : ''}</span>}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
