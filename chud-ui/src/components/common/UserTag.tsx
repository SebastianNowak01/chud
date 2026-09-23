import { Link } from '@tanstack/react-router'
import type { User } from '@/features/users/types'

const TAG = 'inline-flex items-center gap-1.5 font-semibold'

// A username in the user's own color.
export function UserTag({ user }: { user: User | undefined }) {
  if (!user) {
    return <span className={TAG}>nieznany</span>
  }
  return (
    <Link
      to="/users/$userId"
      params={{ userId: user.id }}
      className={`${TAG} no-underline hover:underline`}
      style={{ color: user.color }}
    >
      <span className="size-2.5 flex-none rounded-full" style={{ background: user.color }} />
      {user.username}
    </Link>
  )
}
