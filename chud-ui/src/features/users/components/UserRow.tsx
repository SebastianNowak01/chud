import { useState } from 'react'
import { DrawablyBadge, DrawablyButton } from 'drawably/react'
import { UserTag } from '@/components/common/UserTag'
import { useDeleteUser } from '@/features/users/users-api'
import type { User } from '@/features/users/types'
import { buttonState } from '@/lib/button-state'
import { formatDateTime } from '@/lib/dates'

interface UserRowProps {
  user: User
  onEdit: (user: User) => void
}

export function UserRow({ user, onEdit }: UserRowProps) {
  const [confirmingDelete, setConfirmingDelete] = useState(false)
  const deleteUser = useDeleteUser()

  return (
    <tr>
      <td>
        <UserTag user={user} /> {user.isAdmin && <DrawablyBadge>admin</DrawablyBadge>}
      </td>
      <td>{formatDateTime(user.createdAt)}</td>
      <td className="actions">
        {user.isAdmin ? (
          <span className="hint">Managed via env</span>
        ) : confirmingDelete ? (
          <div className="row">
            {deleteUser.error && <span className="error">{deleteUser.error.message}</span>}
            <span>Delete?</span>
            <DrawablyButton
              tone="danger"
              state={buttonState(deleteUser.status)}
              onClick={() => deleteUser.mutate(user.id)}
            >
              Yes
            </DrawablyButton>
            <DrawablyButton tone="neutral" onClick={() => setConfirmingDelete(false)}>
              No
            </DrawablyButton>
          </div>
        ) : (
          <div className="row">
            <DrawablyButton onClick={() => onEdit(user)}>Edit</DrawablyButton>
            <DrawablyButton tone="danger" onClick={() => setConfirmingDelete(true)}>
              Delete
            </DrawablyButton>
          </div>
        )}
      </td>
    </tr>
  )
}
