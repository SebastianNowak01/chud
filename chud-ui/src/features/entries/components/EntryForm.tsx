import { useState, type FormEvent } from 'react'
import { DrawablyButton, DrawablyCard, DrawablyInput, DrawablyTextarea } from 'drawably/react'
import { useCreateEntry } from '@/features/entries/entries-api'
import { buttonState } from '@/lib/button-state'
import { formatDate, toDateTimeLocal } from '@/lib/dates'

// Set when the entry resolves a planned day.
export interface PlannedDay {
  planId: string
  scheduledFor: string
  excused: boolean
}

interface EntryFormProps {
  activityId: string
  plannedDay?: PlannedDay
  onDone: () => void
}

export function EntryForm({ activityId, plannedDay, onDone }: EntryFormProps) {
  const [description, setDescription] = useState('')
  const [occurredAt, setOccurredAt] = useState(toDateTimeLocal(new Date()))
  const [files, setFiles] = useState<File[]>([])
  const createEntry = useCreateEntry(activityId)
  const excused = plannedDay?.excused ?? false

  const title = !plannedDay
    ? 'Log entry'
    : `${excused ? 'Excuse' : 'Complete'} ${formatDate(plannedDay.scheduledFor)}`

  const submit = (e: FormEvent) => {
    e.preventDefault()
    createEntry.mutate(
      {
        description,
        occurredAt: new Date(occurredAt).toISOString(),
        planId: plannedDay?.planId,
        scheduledFor: plannedDay?.scheduledFor,
        excused,
        files,
      },
      { onSuccess: onDone },
    )
  }

  return (
    <DrawablyCard className="card">
      <form className="stack" onSubmit={submit}>
        <h2>{title}</h2>
        <div className="field">
          <label htmlFor="entry-description">{excused ? 'Why not?' : 'Description'}</label>
          <DrawablyTextarea
            id="entry-description"
            rows={3}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            required={excused}
          />
        </div>
        <div className="field">
          <label htmlFor="entry-occurred-at">When</label>
          <DrawablyInput
            id="entry-occurred-at"
            type="datetime-local"
            value={occurredAt}
            onChange={(e) => setOccurredAt(e.target.value)}
            required
          />
        </div>
        <div className="field">
          <label htmlFor="entry-files">Photos / videos (max 10 MB each)</label>
          <input
            id="entry-files"
            type="file"
            accept="image/*,video/*"
            multiple
            onChange={(e) => setFiles(Array.from(e.target.files ?? []))}
          />
        </div>
        {createEntry.error && <p className="error">{createEntry.error.message}</p>}
        <div className="row">
          <DrawablyButton type="submit" variant="solid" state={buttonState(createEntry.status)}>
            Save
          </DrawablyButton>
          <DrawablyButton type="button" tone="neutral" onClick={onDone}>
            Cancel
          </DrawablyButton>
        </div>
      </form>
    </DrawablyCard>
  )
}
