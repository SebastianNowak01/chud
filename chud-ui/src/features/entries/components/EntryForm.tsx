import { useState, type FormEvent } from 'react'
import { DrawablyButton, DrawablyInput, DrawablyTextarea } from 'drawably/react'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { Field } from '@/components/ui/Field'
import { useCreateEntry } from '@/features/entries/entries-api'
import { buttonState } from '@/lib/button-state'
import { formatDate, toDateString, toDateTimeLocal } from '@/lib/dates'

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
  const now = new Date()
  const excused = plannedDay?.excused ?? false
  const doneDay = plannedDay && !excused ? plannedDay.scheduledFor : null
  const nowLocal = toDateTimeLocal(now)
  const [occurredAt, setOccurredAt] = useState(
    doneDay && doneDay !== toDateString(now) ? `${doneDay}T12:00` : nowLocal,
  )
  const [files, setFiles] = useState<File[]>([])
  const createEntry = useCreateEntry(activityId)

  const title = !plannedDay
    ? 'Nowy wpis'
    : `${excused ? 'Wymówka' : 'Zrobione'}: ${formatDate(plannedDay.scheduledFor)}`
  const minOccurredAt = doneDay ? `${doneDay}T00:00` : undefined
  const maxOccurredAt = doneDay && doneDay < toDateString(now) ? `${doneDay}T23:59` : nowLocal

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
        <Field label={excused ? 'Dlaczego nie?' : 'Opis'} htmlFor="entry-description">
          <DrawablyTextarea
            id="entry-description"
            rows={3}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            required={excused}
          />
        </Field>
        <Field label="Kiedy" htmlFor="entry-occurred-at">
          <DrawablyInput
            id="entry-occurred-at"
            type="datetime-local"
            value={occurredAt}
            min={minOccurredAt}
            max={maxOccurredAt}
            onChange={(e) => setOccurredAt(e.target.value)}
            required
          />
        </Field>
        <Field label="Zdjęcia / filmy (maks. 10 MB każdy)" htmlFor="entry-files">
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
            Zapisz
          </DrawablyButton>
          <DrawablyButton type="button" tone="neutral" onClick={onDone}>
            Anuluj
          </DrawablyButton>
        </div>
      </form>
    </Card>
  )
}
