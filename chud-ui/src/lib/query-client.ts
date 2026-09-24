import { MutationCache, QueryCache, QueryClient } from '@tanstack/react-query'
import { ApiError, errorMessage, isUnauthorized } from '@/lib/api-client'
import { toast } from '@/lib/toast'

declare module '@tanstack/react-query' {
  interface Register {
    mutationMeta: {
      success?: string
      inlineError?: boolean
    }
  }
}

const MAX_RETRIES = 2

const shouldRetry = (failureCount: number, error: unknown) => {
  if (error instanceof ApiError && error.status > 0 && error.status < 500) return false
  return failureCount < MAX_RETRIES
}

export const createQueryClient = () =>
  new QueryClient({
    queryCache: new QueryCache({
      onError: (error) => {
        if (isUnauthorized(error) || (error instanceof ApiError && error.status === 404)) return
        toast.error(errorMessage(error))
      },
    }),
    mutationCache: new MutationCache({
      onError: (error, _variables, _context, mutation) => {
        if (!isUnauthorized(error) && !mutation.meta?.inlineError) toast.error(errorMessage(error))
      },
      onSuccess: (_data, _variables, _context, mutation) => {
        if (mutation.meta?.success) toast.success(mutation.meta.success)
      },
    }),
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        refetchOnWindowFocus: false,
        retry: shouldRetry,
      },
      mutations: {
        retry: false,
      },
    },
  })
