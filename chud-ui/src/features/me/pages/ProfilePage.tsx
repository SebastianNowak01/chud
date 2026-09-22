import { useState, type FormEvent } from 'react'
import { DrawablyButton, DrawablyCard } from 'drawably/react'
import { UserTag } from '@/components/common/UserTag'
import { MyActivityGrid } from '@/features/me/components/MyActivityGrid'
import { useMe, useUpdateMe } from '@/features/me/me-api'
import type { User } from '@/features/users/types'
import { buttonState } from '@/lib/button-state'

export function ProfilePage() {
  const me = useMe()

  return (
    <div className="stack">
      <h2>Profile</h2>
      {me.error && <p className="error">{me.error.message}</p>}
      {me.data && <Profile key={me.data.color} me={me.data} />}
    </div>
  )
}

function Profile({ me }: { me: User }) {
  // The picked color previews live in the grid before it is saved.
  const [color, setColor] = useState(me.color)
  const updateMe = useUpdateMe()

  const submit = (e: FormEvent) => {
    e.preventDefault()
    updateMe.mutate({ color })
  }

  return (
    <>
      <MyActivityGrid me={me} color={color} />
      <DrawablyCard className="card">
        <form className="stack" onSubmit={submit}>
          <div className="field">
            <label htmlFor="color">Your color</label>
            <div className="row">
              <input
                id="color"
                className="color-input"
                type="color"
                value={color}
                onChange={(e) => setColor(e.target.value)}
              />
              <UserTag user={{ ...me, color }} />
            </div>
          </div>
          {updateMe.error && <p className="error">{updateMe.error.message}</p>}
          <div>
            <DrawablyButton type="submit" variant="solid" state={buttonState(updateMe.status)}>
              Save
            </DrawablyButton>
          </div>
        </form>
      </DrawablyCard>
    </>
  )
}
