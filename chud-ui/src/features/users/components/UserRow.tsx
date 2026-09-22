import { useState } from 'react'
import { DrawablyBadge, DrawablyButton } from 'drawably/react'
import { useDeleteUser } from '@/features/users/users-api'
import type { User } from '@/features/users/types'
import { buttonState } from '@/lib/button-state'

interface UserRowProps {
  user: User
  onEdit: (user: User) => void
}

const formatDate = (iso: string) => new Date(iso).toLocaleString()

export function UserRow({ user, onEdit }: UserRowProps) {
  const [confirmingDelete, setConfirmingDelete] = useState(false)
  const deleteUser = useDeleteUser()

  return (
    <tr>
      <td>
        {user.username} {user.isAdmin && <DrawablyBadge>admin</DrawablyBadge>}
      </td>
      <td>{formatDate(user.createdAt)}</td>
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
