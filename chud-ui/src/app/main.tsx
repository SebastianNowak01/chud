import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider, createRouter } from '@tanstack/react-router'
import { QueryClientProvider } from '@tanstack/react-query'
import { ErrorPage } from '@/components/layout/ErrorPage'
import { errorMessage, onUnauthorized, UNEXPECTED_ERROR } from '@/lib/api-client'
import { createQueryClient } from '@/lib/query-client'
import { toast } from '@/lib/toast'

import { routeTree } from '../routeTree.gen'

import '@fontsource/shantell-sans/400.css'
import '@fontsource/shantell-sans/700.css'
import 'drawably/style.css'
import '../tailwind.css'

export const queryClient = createQueryClient()

const router = createRouter({
  routeTree,
  context: {
    queryClient,
  },
  defaultPreload: 'intent',
  defaultErrorComponent: ErrorPage,
  scrollRestoration: true,
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

onUnauthorized(() => {
  queryClient.clear()
  toast.error('Sesja wygasła. Zaloguj się ponownie.', { id: 'session' })
  if (router.state.location.pathname !== '/login') void router.navigate({ to: '/login' })
})

window.addEventListener('unhandledrejection', (event) => {
  console.error(event.reason)
  toast.error(errorMessage(event.reason))
})

window.addEventListener('error', (event) => {
  if (event.message?.includes('ResizeObserver')) return
  console.error(event.error)
  toast.error(UNEXPECTED_ERROR)
})

const rootElement = document.getElementById('app')
if (rootElement && !rootElement.innerHTML) {
  createRoot(rootElement).render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </StrictMode>,
  )
}
