import { Link, Outlet, createRootRouteWithContext } from '@tanstack/react-router'
import type { QueryClient } from '@tanstack/react-query'

export interface RouterContext {
  queryClient: QueryClient
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: Outlet,
  notFoundComponent: () => (
    <div className="mx-auto flex max-w-[960px] flex-col gap-2 px-3 pt-4 pb-12 sm:gap-3 sm:px-4 sm:pt-6 sm:pb-16">
      <h1 className="text-[24px] sm:text-[2em]">Not found</h1>
      <Link to="/">Go home</Link>
    </div>
  ),
})
