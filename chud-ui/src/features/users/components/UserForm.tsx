import { useState, type FormEvent } from 'react'
import { DrawablyButton, DrawablyInput } from 'drawably/react'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { Field } from '@/components/ui/Field'
import { Hint } from '@/components/ui/Hint'
import { useCreateUser, useUpdateUser } from '@/features/users/users-api'
import type { User } from '@/features/users/types'
import { buttonState } from '@/lib/button-state'

interface UserFormProps {
  // The user being edited; omitted when creating a new one.
  user?: User
  onDone: () => void
}

export function UserForm({ user, onDone }: UserFormProps) {
  const isEdit = user !== undefined
  const [username, setUsername] = useState(user?.username ?? '')
  const [password, setPassword] = useState('')

  const createUser = useCreateUser()
  const updateUser = useUpdateUser()
  const mutation = isEdit ? updateUser : createUser

  const submit = (e: FormEvent) => {
    e.preventDefault()
    const options = { onSuccess: onDone }
    if (isEdit) {
      updateUser.mutate({ id: user.id, username, password }, options)
    } else {
      createUser.mutate({ username, password }, options)
    }
  }

  return (
    <Card>
      <form className="flex flex-col gap-4" onSubmit={submit}>
        <h2>{isEdit ? `Edytuj: ${user.username}` : 'Nowy użytkownik'}</h2>
        <Field label="Nazwa użytkownika" htmlFor="user-username">
          <DrawablyInput
            id="user-username"
            autoComplete="off"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            minLength={3}
            maxLength={32}
            required
          />
        </Field>
        <Field label="Hasło" htmlFor="user-password">
          <DrawablyInput
            id="user-password"
            type="password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            minLength={6}
            maxLength={72}
            required={!isEdit}
          />
          {isEdit && <Hint as="span">Zostaw puste, aby nie zmieniać hasła.</Hint>}
        </Field>
        {mutation.error && <ErrorText>{mutation.error.message}</ErrorText>}
        <div className="flex flex-wrap items-center gap-3">
          <DrawablyButton type="submit" variant="solid" state={buttonState(mutation.status)}>
            {isEdit ? 'Zapisz' : 'Utwórz'}
          </DrawablyButton>
          <DrawablyButton type="button" tone="neutral" onClick={onDone}>
            Anuluj
          </DrawablyButton>
        </div>
      </form>
    </Card>
  )
}
