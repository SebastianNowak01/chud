import { useState, type FormEvent } from 'react'
import { DrawablyButton, DrawablyInput, DrawablyTextarea } from 'drawably/react'
import { ErrorText } from '@/components/ui/ErrorText'
import { Field } from '@/components/ui/Field'
import { useCreateEntry } from '@/features/entries/entries-api'
import { mediaFilesError } from '@/features/entries/media'
import { MAX_DESCRIPTION_LENGTH, MEDIA_TYPES } from '@/features/entries/types'
import { buttonState } from '@/lib/button-state'
import { toDateString, toDateTimeLocal } from '@/lib/dates'

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
  const filesError = mediaFilesError(files)
  const createEntry = useCreateEntry(activityId)

  const minOccurredAt = doneDay ? `${doneDay}T00:00` : undefined
  const maxOccurredAt = doneDay && doneDay < toDateString(now) ? `${doneDay}T23:59` : nowLocal

  const submit = (e: FormEvent) => {
    e.preventDefault()
    if (filesError) {
      return
    }
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
    <form className="flex flex-col gap-4" onSubmit={submit}>
      <Field label={excused ? 'Dlaczego nie?' : 'Opis'} htmlFor="entry-description">
        <DrawablyTextarea
          id="entry-description"
          rows={3}
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          maxLength={MAX_DESCRIPTION_LENGTH}
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
          accept={MEDIA_TYPES.join(',')}
          multiple
          onChange={(e) => setFiles(Array.from(e.target.files ?? []))}
        />
      </Field>
      {filesError && <ErrorText>{filesError}</ErrorText>}
      {createEntry.error && <ErrorText>{createEntry.error.message}</ErrorText>}
      <div className="flex flex-wrap items-center gap-3">
        <DrawablyButton
          type="submit"
          variant="solid"
          state={buttonState(createEntry.status)}
          disabled={filesError !== null}
        >
          Zapisz
        </DrawablyButton>
        <DrawablyButton type="button" tone="neutral" onClick={onDone}>
          Anuluj
        </DrawablyButton>
      </div>
    </form>
  )
}
