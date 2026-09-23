import { DrawablySelect } from 'drawably/react'
import { UserTag } from '@/components/common/UserTag'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { Hint } from '@/components/ui/Hint'
import { useLeaderboard } from '@/features/stats/stats-api'
import { useUsersById } from '@/features/users/users-api'
import { GRID_RANGES, type GridRange } from '@/lib/contributions'

const TABLE =
  'w-full border-collapse [&_td]:border-b [&_td]:border-dashed [&_td]:border-rule [&_td]:px-2 [&_td]:py-2 [&_td]:whitespace-nowrap [&_th]:border-b [&_th]:border-dashed [&_th]:border-rule [&_th]:px-2 [&_th]:py-2 [&_th]:text-[13px] [&_th]:font-semibold [&_th]:whitespace-nowrap [&_th]:text-muted'

interface LeaderboardProps {
  from: string
  to: string
  range: GridRange
  onRangeChange: (range: GridRange) => void
}

export function Leaderboard({ from, to, range, onRangeChange }: LeaderboardProps) {
  const leaderboard = useLeaderboard(from, to)
  const usersById = useUsersById()

  return (
    <Card size="lg" className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-4 [&_select]:min-h-11">
        <h3 className="flex-1">Ranking</h3>
        <DrawablySelect aria-label="Zakres" value={range} onChange={(e) => onRangeChange(e.target.value as GridRange)}>
          {GRID_RANGES.map((r) => (
            <option key={r.value} value={r.value}>
              {r.label}
            </option>
          ))}
        </DrawablySelect>
      </div>

      {leaderboard.isPending && <Hint>Ładowanie…</Hint>}
      {leaderboard.error && <ErrorText>{leaderboard.error.message}</ErrorText>}
      {leaderboard.data && (
        <div className="overflow-x-auto">
          <table className={TABLE}>
            <thead>
              <tr>
                <th className="text-right">#</th>
                <th className="text-left">Osoba</th>
                <th className="text-right">Pkt</th>
                <th className="text-right" title="Zrobione zaplanowane dni">✓</th>
                <th className="text-right" title="Wpisy poza planem">+</th>
                <th className="text-right" title="Wymówki">W</th>
                <th className="text-right" title="Opuszczone dni">✗</th>
                <th className="text-right" title="Zrobione / (zrobione + opuszczone)">%</th>
              </tr>
            </thead>
            <tbody>
              {leaderboard.data.map((s, i) => (
                <tr key={s.userId}>
                  <td className="text-right text-muted">{i + 1}</td>
                  <td>
                    <UserTag user={usersById.get(s.userId)} />
                  </td>
                  <td className="text-right font-semibold">{s.points}</td>
                  <td className="text-right">{s.done}</td>
                  <td className="text-right">{s.extra}</td>
                  <td className="text-right">{s.excused}</td>
                  <td className="text-right">{s.missed}</td>
                  <td className="text-right">{s.rate === null ? '–' : `${Math.round(s.rate * 100)}%`}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      <Hint className="m-0">
        ✓ zrobione +3 · + poza planem +1 · W wymówka 0 · ✗ opuszczone −2 · % skuteczność bez wymówek
      </Hint>
    </Card>
  )
}
