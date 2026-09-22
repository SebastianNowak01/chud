import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { authFetch } from '@/lib/api-client'
import type { User } from '@/features/users/types'

const meKey = ['me'] as const

const getMe = async (): Promise<User> => {
  const res = await authFetch('/me')
  return (await res.json()) as User
}

const updateMe = async (payload: { color: string }): Promise<User> => {
  const res = await authFetch('/me', {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
  return (await res.json()) as User
}

export const useMe = () => useQuery({ queryKey: meKey, queryFn: getMe })

export const useUpdateMe = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: updateMe,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: meKey })
      void queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })
}
