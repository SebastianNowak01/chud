import { useState, type FormEvent } from 'react'
import { DrawablyButton, DrawablyInput, DrawablyTextarea } from 'drawably/react'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { Field } from '@/components/ui/Field'
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
    <Card>
      <form className="flex flex-col gap-4" onSubmit={submit}>
        <h2>{title}</h2>
        <Field label={excused ? 'Why not?' : 'Description'} htmlFor="entry-description">
          <DrawablyTextarea
            id="entry-description"
            rows={3}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            required={excused}
          />
        </Field>
        <Field label="When" htmlFor="entry-occurred-at">
          <DrawablyInput
            id="entry-occurred-at"
            type="datetime-local"
            value={occurredAt}
            onChange={(e) => setOccurredAt(e.target.value)}
            required
          />
        </Field>
        <Field label="Photos / videos (max 10 MB each)" htmlFor="entry-files">
          <input
            id="entry-files"
            type="file"
            accept="image/*,video/*"
            multiple
            onChange={(e) => setFiles(Array.from(e.target.files ?? []))}
          />
        </Field>
        {createEntry.error && <ErrorText>{createEntry.error.message}</ErrorText>}
        <div className="flex flex-wrap items-center gap-3">
          <DrawablyButton type="submit" variant="solid" state={buttonState(createEntry.status)}>
            Save
          </DrawablyButton>
          <DrawablyButton type="button" tone="neutral" onClick={onDone}>
            Cancel
          </DrawablyButton>
        </div>
      </form>
    </Card>
  )
}
