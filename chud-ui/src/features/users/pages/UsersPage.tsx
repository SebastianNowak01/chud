import { useState } from 'react'
import { DrawablyButton, DrawablyCard } from 'drawably/react'
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
    <div className="stack">
      <div className="header">
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

      <DrawablyCard className="card">
        {users.isPending && <p className="hint">Loading…</p>}
        {users.error && <p className="error">{users.error.message}</p>}
        {users.data && (
          <div className="table-wrap">
            <table>
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
      </DrawablyCard>
    </div>
  )
}
