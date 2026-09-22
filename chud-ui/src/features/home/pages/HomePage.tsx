import { getRouteApi } from '@tanstack/react-router'

const authenticatedRoute = getRouteApi('/_authenticated')

export function HomePage() {
  const { session } = authenticatedRoute.useRouteContext()
  return <h2>Hello, {session.username}</h2>
}
