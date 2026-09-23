import { useState, type FormEvent } from 'react'
import { DrawablyButton, DrawablyUnderline } from 'drawably/react'
import { ColorSwatch } from '@/components/common/ColorSwatch'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { Hint } from '@/components/ui/Hint'
import { MyActivityGrid } from '@/features/me/components/MyActivityGrid'
import { useMe, useUpdateMe } from '@/features/me/me-api'
import type { User } from '@/features/users/types'
import { buttonState } from '@/lib/button-state'
import { LOCALE } from '@/lib/dates'

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
  const changed = color !== me.color
  const memberSince = new Date(me.createdAt).toLocaleDateString(LOCALE, { month: 'long', year: 'numeric' })

  const submit = (e: FormEvent) => {
    e.preventDefault()
    updateMe.mutate({ color })
  }

  return (
    <>
      <Card className="w-fit max-w-full">
        <form className="flex flex-wrap items-center gap-6" onSubmit={submit}>
          <label className="group flex cursor-pointer flex-col items-center gap-1" title="Pick your color">
            <ColorSwatch
              color={color}
              className="size-[104px] rounded-full transition-transform duration-150 ease-[ease] group-hover:scale-[1.04] group-hover:-rotate-6 group-focus-within:outline-2 group-focus-within:outline-offset-4 group-focus-within:outline-ink group-focus-within:outline-dashed"
            />
            <input
              type="color"
              className="sr-only"
              aria-label="Your color"
              value={color}
              onChange={(e) => setColor(e.target.value)}
            />
            <Hint as="span">tap to change</Hint>
          </label>

          <div className="flex min-w-0 flex-col items-start gap-1.5">
            <h2 className="text-[32px] wrap-anywhere" style={{ color }}>
              <DrawablyUnderline stroke={color}>{me.username}</DrawablyUnderline>
            </h2>
            <Hint as="span">
              {me.isAdmin ? 'Admin · ' : ''}member since {memberSince}
            </Hint>
            <code className="text-[14px] text-muted">{color}</code>
            {updateMe.error && <ErrorText>{updateMe.error.message}</ErrorText>}
            {changed && (
              <div className="flex flex-wrap items-center gap-3">
                <DrawablyButton type="submit" variant="solid" state={buttonState(updateMe.status)}>
                  Save color
                </DrawablyButton>
                <DrawablyButton type="button" tone="neutral" onClick={() => setColor(me.color)}>
                  Undo
                </DrawablyButton>
              </div>
            )}
          </div>
        </form>
      </Card>

      <MyActivityGrid me={me} color={color} />
    </>
  )
}
