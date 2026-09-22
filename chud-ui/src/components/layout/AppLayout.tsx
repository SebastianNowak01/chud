import { getRouteApi, Outlet, useNavigate } from '@tanstack/react-router'
import { useQueryClient } from '@tanstack/react-query'
import { DrawablyButton, DrawablyDivider } from 'drawably/react'
import { clearToken } from '@/features/auth/lib/token'

const authenticatedRoute = getRouteApi('/_authenticated')

export function AppLayout() {
  const { session } = authenticatedRoute.useRouteContext()
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
        <h1>chud</h1>
        <div className="header__user">
          <span>
            {session.username}
            {session.is_admin && ' (admin)'}
          </span>
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
