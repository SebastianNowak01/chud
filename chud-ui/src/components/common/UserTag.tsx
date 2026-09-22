import type { User } from '@/features/users/types'

// A username in the user's own color.
export function UserTag({ user }: { user: User | undefined }) {
  if (!user) {
    return <span className="user-tag">unknown</span>
  }
  return (
    <span className="user-tag" style={{ color: user.color }}>
      <span className="user-tag__dot" style={{ background: user.color }} />
      {user.username}
    </span>
  )
}
