import { useState } from 'react'
import { DrawablyButton } from 'drawably/react'
import { UserTag } from '@/components/common/UserTag'
import { Badge } from '@/components/ui/Badge'
import { ErrorText } from '@/components/ui/ErrorText'
import { Hint } from '@/components/ui/Hint'
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
        <UserTag user={user} /> {user.isAdmin && <Badge tone="accent">admin</Badge>}
      </td>
      <td>{formatDateTime(user.createdAt)}</td>
      <td className="text-right">
        {user.isAdmin ? (
          <Hint as="span">Zarządzany przez env</Hint>
        ) : confirmingDelete ? (
          <div className="flex flex-wrap items-center gap-3 justify-end">
            {deleteUser.error && <ErrorText as="span">{deleteUser.error.message}</ErrorText>}
            <span>Usunąć?</span>
            <DrawablyButton
              tone="danger"
              state={buttonState(deleteUser.status)}
              onClick={() => deleteUser.mutate(user.id)}
            >
              Tak
            </DrawablyButton>
            <DrawablyButton tone="neutral" onClick={() => setConfirmingDelete(false)}>
              Nie
            </DrawablyButton>
          </div>
        ) : (
          <div className="flex flex-wrap items-center gap-3 justify-end">
            <DrawablyButton onClick={() => onEdit(user)}>Edytuj</DrawablyButton>
            <DrawablyButton tone="danger" onClick={() => setConfirmingDelete(true)}>
              Usuń
            </DrawablyButton>
          </div>
        )}
      </td>
    </tr>
  )
}
