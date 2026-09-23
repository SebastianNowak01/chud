import { useState, type FormEvent } from 'react'
import { DrawablyButton, DrawablyCard, DrawablyUnderline } from 'drawably/react'
import { ColorSwatch } from '@/components/common/ColorSwatch'
import { MyActivityGrid } from '@/features/me/components/MyActivityGrid'
import { useMe, useUpdateMe } from '@/features/me/me-api'
import type { User } from '@/features/users/types'
import { buttonState } from '@/lib/button-state'
import { LOCALE } from '@/lib/dates'

export function ProfilePage() {
  const me = useMe()

  return (
    <div className="stack">
      {me.error && <p className="error">{me.error.message}</p>}
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
      <DrawablyCard className="card profile-card">
        <form className="profile-card__body" onSubmit={submit}>
          <label className="profile-card__swatch" title="Pick your color">
            <ColorSwatch color={color} />
            <input
              type="color"
              className="visually-hidden"
              aria-label="Your color"
              value={color}
              onChange={(e) => setColor(e.target.value)}
            />
            <span className="hint">tap to change</span>
          </label>

          <div className="profile-card__info">
            <h2 style={{ color }}>
              <DrawablyUnderline stroke={color}>{me.username}</DrawablyUnderline>
            </h2>
            <span className="hint">
              {me.isAdmin ? 'Admin · ' : ''}member since {memberSince}
            </span>
            <code className="profile-card__hex">{color}</code>
            {updateMe.error && <p className="error">{updateMe.error.message}</p>}
            {changed && (
              <div className="row">
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
      </DrawablyCard>

      <MyActivityGrid me={me} color={color} />
    </>
  )
}
