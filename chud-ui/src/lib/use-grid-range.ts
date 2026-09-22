import { useMemo } from 'react'
import { GRID_RANGES, gridLayout, type GridRange } from '@/lib/contributions'
import { toDateString } from '@/lib/dates'
import { useStoredState } from '@/lib/use-stored-state'

const RANGE_VALUES = GRID_RANGES.map((r) => r.value)

// The selected grid range (remembered per grid) and the weeks it covers; defaults to a month.
export const useGridRange = (storageKey: string) => {
  const [range, setRange] = useStoredState<GridRange>(storageKey, 'month', RANGE_VALUES)
  const today = toDateString(new Date())
  // eslint-disable-next-line react-hooks/exhaustive-deps -- recompute when the day changes, not every render
  const layout = useMemo(() => gridLayout(range), [range, today])
  return { range, setRange, layout }
}
