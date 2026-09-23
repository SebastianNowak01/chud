import type { User } from '@/features/users/types'

const TAG = 'inline-flex items-center gap-1.5 font-semibold'

// A username in the user's own color.
export function UserTag({ user }: { user: User | undefined }) {
  if (!user) {
    return <span className={TAG}>unknown</span>
  }
  return (
    <span className={TAG} style={{ color: user.color }}>
      <span className="size-2.5 flex-none rounded-full" style={{ background: user.color }} />
      {user.username}
    </span>
  )
}
