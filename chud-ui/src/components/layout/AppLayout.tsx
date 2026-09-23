import { getRouteApi, Link, Outlet, useNavigate } from '@tanstack/react-router'
import { useQueryClient } from '@tanstack/react-query'
import { DrawablyButton, DrawablyDivider } from 'drawably/react'
import { UserTag } from '@/components/common/UserTag'
import { Logo } from '@/components/layout/Logo'
import { clearToken } from '@/features/auth/lib/token'
import { useMe } from '@/features/me/me-api'

const authenticatedRoute = getRouteApi('/_authenticated')

const NAV_LINK = {
  className: 'inline-flex min-h-11 items-center no-underline sm:inline sm:min-h-0',
  activeProps: { className: 'font-semibold text-ink' },
  inactiveProps: { className: 'text-muted' },
}

export function AppLayout() {
  const { session } = authenticatedRoute.useRouteContext()
  const me = useMe()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const logout = () => {
    clearToken()
    queryClient.clear()
    void navigate({ to: '/login' })
  }

  return (
    <div className="mx-auto flex max-w-[960px] flex-col gap-2 px-3 pt-4 pb-12 sm:gap-3 sm:px-4 sm:pt-6 sm:pb-16">
      <header className="flex flex-wrap items-center justify-between gap-4">
        <nav className="flex flex-wrap items-center gap-3.5 sm:gap-5">
          <Logo size="small" linked />
          <Link to="/activities" {...NAV_LINK}>
            Activities
          </Link>
          <Link to="/profile" {...NAV_LINK}>
            Profile
          </Link>
          {session.is_admin && (
            <Link to="/users" {...NAV_LINK}>
              Users
            </Link>
          )}
        </nav>
        <div className="flex items-center gap-3 text-muted">
          {me.data ? <UserTag user={me.data} /> : <span>{session.username}</span>}
          <DrawablyButton tone="neutral" onClick={logout}>
            Log out
          </DrawablyButton>
        </div>
      </header>
      <DrawablyDivider />
      <Outlet />
    </div>
  )
}
