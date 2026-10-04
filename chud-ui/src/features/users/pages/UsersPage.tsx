import { useState } from 'react'
import { DrawablyButton } from 'drawably/react'
import { Card } from '@/components/ui/Card'
import { LoadError } from '@/components/ui/LoadError'
import { Hint } from '@/components/ui/Hint'
import { Modal } from '@/components/ui/Modal'
import { UserForm } from '@/features/users/components/UserForm'
import { UserRow } from '@/features/users/components/UserRow'
import { useUsers } from '@/features/users/users-api'
import type { User } from '@/features/users/types'
import { usePageTitle } from '@/lib/use-page-title'

type FormState = { mode: 'closed' } | { mode: 'create' } | { mode: 'edit'; user: User }

export function UsersPage() {
  usePageTitle('Użytkownicy')
  const users = useUsers()
  const [form, setForm] = useState<FormState>({ mode: 'closed' })
  const closeForm = () => setForm({ mode: 'closed' })

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <h2>Użytkownicy</h2>
        <DrawablyButton variant="solid" onClick={() => setForm({ mode: 'create' })}>
          Dodaj użytkownika
        </DrawablyButton>
      </div>

      {form.mode !== 'closed' && (
        <Modal title={form.mode === 'edit' ? `Edytuj: ${form.user.username}` : 'Nowy użytkownik'} onClose={closeForm}>
          <UserForm user={form.mode === 'edit' ? form.user : undefined} onDone={closeForm} />
        </Modal>
      )}

      <Card>
        {users.isPending && <Hint>Ładowanie…</Hint>}
        <LoadError error={users.error} onRetry={() => void users.refetch()} />
        {users.data && (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse [&_td]:border-b [&_td]:border-dashed [&_td]:border-rule [&_td]:px-2 [&_td]:py-2.5 [&_td]:text-left [&_td]:whitespace-nowrap [&_th]:border-b [&_th]:border-dashed [&_th]:border-rule [&_th]:px-2 [&_th]:py-2.5 [&_th]:text-left [&_th]:text-[13px] [&_th]:font-semibold [&_th]:whitespace-nowrap [&_th]:text-muted">
              <thead>
                <tr>
                  <th>Nazwa</th>
                  <th>Utworzono</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {users.data.map((user) => (
                  <UserRow key={user.id} user={user} onEdit={(u) => setForm({ mode: 'edit', user: u })} />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </div>
  )
}
