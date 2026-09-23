import { useState } from 'react'
import { DrawablyButton } from 'drawably/react'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { Hint } from '@/components/ui/Hint'
import { UserForm } from '@/features/users/components/UserForm'
import { UserRow } from '@/features/users/components/UserRow'
import { useUsers } from '@/features/users/users-api'
import type { User } from '@/features/users/types'

type FormState = { mode: 'closed' } | { mode: 'create' } | { mode: 'edit'; user: User }

export function UsersPage() {
  const users = useUsers()
  const [form, setForm] = useState<FormState>({ mode: 'closed' })
  const closeForm = () => setForm({ mode: 'closed' })

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <h2>Users</h2>
        {form.mode === 'closed' && (
          <DrawablyButton variant="solid" onClick={() => setForm({ mode: 'create' })}>
            Add user
          </DrawablyButton>
        )}
      </div>

      {form.mode !== 'closed' && (
        <UserForm
          key={form.mode === 'edit' ? form.user.id : 'new'}
          user={form.mode === 'edit' ? form.user : undefined}
          onDone={closeForm}
        />
      )}

      <Card>
        {users.isPending && <Hint>Loading…</Hint>}
        {users.error && <ErrorText>{users.error.message}</ErrorText>}
        {users.data && (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse [&_td]:border-b [&_td]:border-dashed [&_td]:border-rule [&_td]:px-2 [&_td]:py-2.5 [&_td]:text-left [&_td]:whitespace-nowrap [&_th]:border-b [&_th]:border-dashed [&_th]:border-rule [&_th]:px-2 [&_th]:py-2.5 [&_th]:text-left [&_th]:text-[13px] [&_th]:font-semibold [&_th]:whitespace-nowrap [&_th]:text-muted">
              <thead>
                <tr>
                  <th>Username</th>
                  <th>Created</th>
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
