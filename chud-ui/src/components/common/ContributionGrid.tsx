import { useEffect, useMemo, useRef, useState, type KeyboardEvent, type ReactNode, type SyntheticEvent } from 'react'
import { DrawablySelect } from 'drawably/react'
import { roughRoundedRect, scribbleFill } from 'drawably'
import { UserTag } from '@/components/common/UserTag'
import { Card, type CardSize } from '@/components/ui/Card'
import { LoadError } from '@/components/ui/LoadError'
import { Hint } from '@/components/ui/Hint'
import type { Activity } from '@/features/activities/types'
import type { Entry } from '@/features/entries/types'
import type { User } from '@/features/users/types'
import {
  GRID_RANGES,
  groupByDay,
  level,
  peopleOnDay,
  scaledLevel,
  type Day,
  type GridLayout,
  type GridRange,
} from '@/lib/contributions'
import { formatDate, formatMonth, formatTime, toDateString } from '@/lib/dates'

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
const DAY_LABELS = ['Pn', '', 'Śr', '', 'Pt', '', '']

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
  level: number // 0 = only excused
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
        const lvl = stripe.level
        return lvl > 0 ? (
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
  onRetry?: () => void
  coloring: GridColoring
  usersById: Map<string, User>
  activitiesById: Map<string, Activity>
  size?: CardSize
  weekStatus?: WeekStatus
}

export interface WeekStatus {
  onFocusDate: (date: string) => void
}

export function ContributionGrid({
  title,
  layout,
  range,
  onRangeChange,
  entries,
  error,
  onRetry,
  coloring,
  usersById,
  activitiesById,
  size,
  weekStatus,
}: ContributionGridProps) {
  const today = toDateString(new Date())
  const [selected, setSelected] = useState(today)
  const days = useMemo(() => groupByDay(entries ?? []), [entries])
  const maxDone = useMemo(
    () => Math.max(0, ...layout.weeks.flat().map((date) => days.get(date)?.done ?? 0)),
    [days, layout],
  )
  const wrapper = useRef<HTMLDivElement>(null)
  const [hovered, setHovered] = useState<Hovered | null>(null)
  const [pointed, setPointed] = useState<string | null>(null)
  const focused = pointed ?? selected
  const focusedCol = weekStatus ? layout.weeks.findIndex((week) => week.includes(focused)) : -1
  const timer = useRef<number | undefined>(undefined)
  const open = useRef(false)

  useEffect(() => () => window.clearTimeout(timer.current), [])

  const onFocusDate = weekStatus?.onFocusDate
  useEffect(() => onFocusDate?.(focused), [onFocusDate, focused])

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
      return [{ color: coloring.color, level: scaledLevel(day.done, maxDone) }]
    }
    // In the small year view stripes would be unreadable, so only the day's leader is drawn.
    const people = range === 'year' ? peopleOnDay(day).slice(0, 1) : peopleOnDay(day)
    return people.map((p) => ({
      color: usersById.get(p.userId)?.color ?? '#888888',
      level: p.excusedOnly ? 0 : level(day.done),
    }))
  }

  const select = (date: string) => (e?: KeyboardEvent) => {
    if (e && e.key !== 'Enter' && e.key !== ' ') return
    e?.preventDefault()
    setSelected(date)
  }

  const show = (next: Hovered | null) => {
    open.current = next !== null
    setHovered(next)
  }

  const hover = (date: string, instant: boolean) => (e: SyntheticEvent<SVGRectElement>) => {
    const box = wrapper.current?.getBoundingClientRect()
    if (!box) return
    const cell = e.currentTarget.getBoundingClientRect()
    const x = cell.left + cell.width / 2 - box.left
    const align = x < POPOVER_EDGE ? 'start' : x > box.width - POPOVER_EDGE ? 'end' : 'center'
    const next: Hovered = { date, x, y: cell.top - box.top, align }
    window.clearTimeout(timer.current)
    setPointed(date)
    if (instant || open.current) {
      show(next)
    } else {
      timer.current = window.setTimeout(() => show(next), POPOVER_OPEN_DELAY)
    }
  }

  const unhover = () => {
    window.clearTimeout(timer.current)
    timer.current = window.setTimeout(() => {
      show(null)
      setPointed(null)
    }, POPOVER_CLOSE_DELAY)
  }
  const showActivities = coloring.kind === 'single'

  return (
    <Card size={size} className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-4 [&_select]:min-h-11">
        <h3 className="flex-1">{title}</h3>
        <DrawablySelect
          aria-label="Zakres"
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

      <LoadError error={error} onRetry={onRetry} />

      <div ref={wrapper} className="relative">
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

          {focusedCol >= 0 && (
            <rect
              x={LABEL_WIDTH + focusedCol * STEP - 2}
              y={HEADER_HEIGHT - 2}
              width={CELL + 4}
              height={6 * STEP + CELL + 4}
              rx={6}
              className="fill-ink/8"
            />
          )}
          {layout.weeks.map((week, col) =>
            week.map((date, weekday) => {
              const day = days.get(date)
              const isFuture = date > today
              const summary = day ? `zrobione: ${day.done}${day.excused ? `, wymówki: ${day.excused}` : ''}` : 'nic'

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
                      onPointerEnter={hover(date, false)}
                      onPointerLeave={unhover}
                      onFocus={hover(date, true)}
                      onBlur={unhover}
                    />
                  )}
                </g>
              )
            }),
          )}
        </svg>
        {hovered && (
          <Popover hovered={hovered}>
            <p className="m-0 font-semibold">{formatDate(hovered.date)}</p>
            <DayPeople
              day={days.get(hovered.date)}
              usersById={usersById}
              activitiesById={activitiesById}
              showActivities={showActivities}
              compact
            />
          </Popover>
        )}
      </div>

      <Legend />

      <p className="m-0 font-semibold">{formatDate(selected)}</p>
      <DayPeople
        day={days.get(selected)}
        usersById={usersById}
        activitiesById={activitiesById}
        showActivities={showActivities}
      />
    </Card>
  )
}

function Legend() {
  const samples: Stripe[][] = [[], ...[1, 2, 3, 4].map((lvl) => [{ color: 'var(--color-ink)', level: lvl }])]
  const legendCell = (key: string, stripes: Stripe[], i: number) => (
    <g key={key} transform={`translate(${i * STEP} 0)`}>
      <SketchCell seedKey={key} stripes={stripes} />
    </g>
  )

  return (
    <div className="flex flex-wrap items-center gap-1 text-[11px] text-muted">
      <span>mniej</span>
      <svg viewBox={`0 0 ${5 * STEP - 4} ${CELL}`} className="h-[11px] w-auto">
        {samples.map((stripes, i) => legendCell(`legend-${i}`, stripes, i))}
      </svg>
      <span>więcej</span>
      <svg viewBox={`0 0 ${CELL} ${CELL}`} className="ml-2 size-[11px]">
        <SketchCell seedKey="legend-excused" stripes={[{ color: 'var(--color-ink)', level: 0 }]} />
      </svg>
      <span>wymówka</span>
    </div>
  )
}

interface Hovered {
  date: string
  x: number
  y: number
  align: 'start' | 'center' | 'end'
}

const POPOVER_EDGE = 130
const POPOVER_OPEN_DELAY = 500
const POPOVER_CLOSE_DELAY = 150

const POPOVER_ALIGN = {
  start: '',
  center: '-translate-x-1/2',
  end: '',
}

function Popover({ hovered, children }: { hovered: Hovered; children: ReactNode }) {
  const { x, y, align } = hovered
  const position = align === 'start' ? { left: 0 } : align === 'end' ? { right: 0 } : { left: x }
  return (
    <div
      role="tooltip"
      className={`pointer-events-none absolute z-10 w-max max-w-[260px] -translate-y-full ${POPOVER_ALIGN[align]}`}
      style={{ ...position, top: y - 6 }}
    >
      <Card filled className="flex flex-col gap-1">{children}</Card>
    </div>
  )
}

interface DayPeopleProps {
  day: Day | undefined
  usersById: Map<string, User>
  activitiesById: Map<string, Activity>
  showActivities: boolean
  compact?: boolean
}

function DayPeople({ day, usersById, activitiesById, showActivities, compact = false }: DayPeopleProps) {
  if (!day?.entries.length) {
    return <Hint className="m-0">Nic nie zapisano.</Hint>
  }
  return (
    <ul className="m-0 list-none p-0">
      {day.entries.map((entry) => (
        <li
          key={entry.id}
          className={`flex flex-wrap items-center gap-x-2 border-t border-dashed border-rule ${compact ? 'py-1' : 'min-h-11 py-1.5'}`}
        >
          <span className="text-[13px] text-muted tabular-nums">{formatTime(entry.occurredAt)}</span>
          <UserTag user={usersById.get(entry.userId)} />
          {showActivities && <span>{activitiesById.get(entry.activityId)?.name ?? 'aktywność'}</span>}
          {entry.excused && <Hint as="span">wymówka</Hint>}
          {entry.description && (
            <Hint as="span" className={compact ? 'w-full truncate' : 'w-full sm:w-auto'}>
              {entry.description}
            </Hint>
          )}
        </li>
      ))}
    </ul>
  )
}
