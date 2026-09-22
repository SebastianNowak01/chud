import { useState, type FormEvent } from 'react'
import { DrawablyButton, DrawablyCard, DrawablyInput } from 'drawably/react'
import { useCreateActivity } from '@/features/activities/activities-api'
import { buttonState } from '@/lib/button-state'

export function ActivityForm({ onDone }: { onDone: () => void }) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const createActivity = useCreateActivity()

  const submit = (e: FormEvent) => {
    e.preventDefault()
    createActivity.mutate({ name, description }, { onSuccess: onDone })
  }

  return (
    <DrawablyCard className="card">
      <form className="stack" onSubmit={submit}>
        <h2>New activity</h2>
        <div className="field">
          <label htmlFor="activity-name">Name</label>
          <DrawablyInput
            id="activity-name"
            placeholder="gym, laundry, walk…"
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={50}
            required
          />
        </div>
        <div className="field">
          <label htmlFor="activity-description">Description</label>
          <DrawablyInput
            id="activity-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>
        {createActivity.error && <p className="error">{createActivity.error.message}</p>}
        <div className="row">
          <DrawablyButton type="submit" variant="solid" state={buttonState(createActivity.status)}>
            Create
          </DrawablyButton>
          <DrawablyButton type="button" tone="neutral" onClick={onDone}>
            Cancel
          </DrawablyButton>
        </div>
      </form>
    </DrawablyCard>
  )
}
