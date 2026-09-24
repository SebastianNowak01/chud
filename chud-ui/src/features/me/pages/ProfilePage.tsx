import { useState, type FormEvent } from 'react'
import { DrawablyButton, DrawablyUnderline } from 'drawably/react'
import { ColorSwatch } from '@/components/common/ColorSwatch'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { Hint } from '@/components/ui/Hint'
import { MyActivityGrid } from '@/features/me/components/MyActivityGrid'
import { useMe, useUpdateMe } from '@/features/me/me-api'
import { WeekMood } from '@/features/stats/components/WeekMood'
import { useWeekStatus } from '@/features/stats/use-week-status'
import type { User } from '@/features/users/types'
import { buttonState } from '@/lib/button-state'
import { LOCALE } from '@/lib/dates'
import { useGridRange } from '@/lib/use-grid-range'

export function ProfilePage() {
  const me = useMe()

  return (
    <div className="flex flex-col gap-4">
      {me.error && <ErrorText>{me.error.message}</ErrorText>}
      {me.data && <Profile key={me.data.color} me={me.data} />}
    </div>
  )
}

function Profile({ me }: { me: User }) {
  // The picked color previews live in the card and the grid before it is saved.
  const [color, setColor] = useState(me.color)
  const updateMe = useUpdateMe()
  const { range, setRange, layout } = useGridRange('grid:me')
  const { week, counts, loaded, error, gridStatus } = useWeekStatus(me.id, layout.firstDay)
  const changed = color !== me.color
  const memberSince = new Date(me.createdAt).toLocaleDateString(LOCALE, { day: 'numeric', month: 'long', year: 'numeric' })

  const submit = (e: FormEvent) => {
    e.preventDefault()
    updateMe.mutate({ color })
  }

  return (
    <>
      <div className="grid items-stretch gap-4 lg:grid-cols-2">
        <Card className="flex items-center">
          <form className="flex flex-wrap items-center gap-6" onSubmit={submit}>
            <label className="group flex cursor-pointer flex-col items-center gap-1" title="Wybierz swój kolor">
              <ColorSwatch
                color={color}
                className="size-[104px] rounded-full transition-transform duration-150 ease-[ease] group-hover:scale-[1.04] group-hover:-rotate-6 group-focus-within:outline-2 group-focus-within:outline-offset-4 group-focus-within:outline-ink group-focus-within:outline-dashed"
              />
              <input
                type="color"
                className="sr-only"
                aria-label="Twój kolor"
                value={color}
                onChange={(e) => setColor(e.target.value)}
              />
              <Hint as="span">kliknij, aby zmienić</Hint>
            </label>

            <div className="flex min-w-0 flex-col items-start gap-1.5">
              <h2 className="text-[32px] wrap-anywhere" style={{ color }}>
                <DrawablyUnderline stroke={color}>{me.username}</DrawablyUnderline>
              </h2>
              <Hint as="span">
                {me.isAdmin ? 'Admin · ' : ''}w grupie od {memberSince}
              </Hint>
              <code className="text-[14px] text-muted">{color}</code>
              {updateMe.error && <ErrorText>{updateMe.error.message}</ErrorText>}
              {changed && (
                <div className="flex flex-wrap items-center gap-3">
                  <DrawablyButton type="submit" variant="solid" state={buttonState(updateMe.status)}>
                    Zapisz kolor
                  </DrawablyButton>
                  <DrawablyButton type="button" tone="neutral" onClick={() => setColor(me.color)}>
                    Cofnij
                  </DrawablyButton>
                </div>
              )}
            </div>
          </form>
        </Card>
        <Card className="flex items-center">
          <WeekMood week={week} counts={counts} loaded={loaded} error={error} />
        </Card>
      </div>

      <MyActivityGrid
        me={me}
        color={color}
        range={range}
        onRangeChange={setRange}
        layout={layout}
        weekStatus={gridStatus}
      />
    </>
  )
}
