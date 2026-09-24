import { useState } from 'react'
import { getRouteApi } from '@tanstack/react-router'
import { DrawablyUnderline } from 'drawably/react'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { LoadError } from '@/components/ui/LoadError'
import { Hint } from '@/components/ui/Hint'
import { ColorCard } from '@/features/me/components/ColorCard'
import { WeekMood } from '@/features/stats/components/WeekMood'
import { useLeaderboard } from '@/features/stats/stats-api'
import { useWeekStatus } from '@/features/stats/use-week-status'
import { UserActivityGrid } from '@/features/users/components/UserActivityGrid'
import { useUsers } from '@/features/users/users-api'
import { LOCALE } from '@/lib/dates'
import { useGridRange } from '@/lib/use-grid-range'

const userRoute = getRouteApi('/_authenticated/users/$userId')

export function UserProfilePage() {
  const { userId } = userRoute.useParams()
  const { session } = userRoute.useRouteContext()
  const users = useUsers()
  const { range, setRange, layout } = useGridRange('grid:user')
  const leaderboard = useLeaderboard(layout.firstDay, layout.lastDay)
  const weekStatus = useWeekStatus(userId, layout.firstDay)
  const [pickedColor, setPickedColor] = useState<string | null>(null)

  const user = users.data?.find((u) => u.id === userId)
  const place = leaderboard.data?.findIndex((s) => s.userId === userId) ?? -1
  const stats = place >= 0 ? leaderboard.data?.[place] : undefined

  if (users.error) {
    return <LoadError error={users.error} onRetry={() => void users.refetch()} />
  }
  if (users.data && !user) {
    return <ErrorText>Nie znaleziono użytkownika.</ErrorText>
  }
  if (!user) {
    return <Hint>Ładowanie…</Hint>
  }

  const isMe = session.user_id === user.id
  const color = (isMe && pickedColor) || user.color
  const memberSince = new Date(user.createdAt).toLocaleDateString(LOCALE, {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
  const tiles = stats
    ? [
        { label: 'miejsce', value: `${place + 1}.` },
        { label: 'punkty', value: stats.points },
        { label: 'zrobione', value: stats.done },
        { label: 'poza planem', value: stats.extra },
        { label: 'wymówki', value: stats.excused },
        { label: 'opuszczone', value: stats.missed },
        { label: 'skuteczność', value: stats.rate === null ? '–' : `${Math.round(stats.rate * 100)}%` },
      ]
    : []

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-col gap-1.5">
          <h2 className="text-[32px] wrap-anywhere" style={{ color }}>
            <DrawablyUnderline stroke={color}>{user.username}</DrawablyUnderline>
          </h2>
          <Hint as="span">
            {user.isAdmin ? 'Admin · ' : ''}w grupie od {memberSince}
          </Hint>
        </div>
      </div>

      <div className={`grid items-stretch gap-4 ${isMe ? 'lg:grid-cols-2' : ''}`}>
        {isMe && <ColorCard color={color} savedColor={user.color} onColorChange={setPickedColor} />}
        <Card size={isMe ? 'md' : 'lg'} className="flex items-center">
          <WeekMood
            week={weekStatus.week}
            counts={weekStatus.counts}
            loaded={weekStatus.loaded}
            error={weekStatus.error}
            onRetry={weekStatus.retry}
          />
        </Card>
      </div>

      <Card size="lg" className="flex flex-col gap-3">
        <h3>Statystyki</h3>
        <LoadError error={leaderboard.error} onRetry={() => void leaderboard.refetch()} />
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4 lg:grid-cols-7">
          {tiles.map((tile) => (
            <div key={tile.label} className="flex flex-col">
              <span className="text-[26px] font-semibold">{tile.value}</span>
              <Hint as="span">{tile.label}</Hint>
            </div>
          ))}
        </div>
      </Card>

      <UserActivityGrid
        user={user}
        color={color}
        range={range}
        onRangeChange={setRange}
        layout={layout}
        weekStatus={weekStatus.gridStatus}
      />
    </div>
  )
}
