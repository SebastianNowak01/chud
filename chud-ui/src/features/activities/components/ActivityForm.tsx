import { useState, type FormEvent } from 'react'
import { DrawablyButton, DrawablyInput } from 'drawably/react'
import { FormError } from '@/components/ui/FormError'
import { Field } from '@/components/ui/Field'
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
    <form className="flex flex-col gap-4" onSubmit={submit}>
      <Field label="Nazwa" htmlFor="activity-name">
        <DrawablyInput
          id="activity-name"
          placeholder="siłownia, pranie, spacer…"
          value={name}
          onChange={(e) => setName(e.target.value)}
          maxLength={50}
          required
        />
      </Field>
      <Field label="Opis" htmlFor="activity-description">
        <DrawablyInput
          id="activity-description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          maxLength={500}
        />
      </Field>
      <FormError error={createActivity.error} />
      <div className="flex flex-wrap items-center gap-3">
        <DrawablyButton type="submit" variant="solid" state={buttonState(createActivity.status)}>
          Utwórz
        </DrawablyButton>
        <DrawablyButton type="button" tone="neutral" onClick={onDone}>
          Anuluj
        </DrawablyButton>
      </div>
    </form>
  )
}
