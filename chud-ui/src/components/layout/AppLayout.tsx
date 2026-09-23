import { getRouteApi, Link, Outlet, useNavigate } from '@tanstack/react-router'
import { useQueryClient } from '@tanstack/react-query'
import { DrawablyButton, DrawablyDivider } from 'drawably/react'
import { UserTag } from '@/components/common/UserTag'
import { Logo } from '@/components/layout/Logo'
import { clearToken } from '@/features/auth/lib/token'
import { useMe } from '@/features/me/me-api'

const authenticatedRoute = getRouteApi('/_authenticated')

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
    <div className="page">
      <header className="header">
        <nav className="row nav">
          <Logo size="small" linked />
          <Link to="/activities">Activities</Link>
          <Link to="/profile">Profile</Link>
          {session.is_admin && <Link to="/users">Users</Link>}
        </nav>
        <div className="header__user">
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
