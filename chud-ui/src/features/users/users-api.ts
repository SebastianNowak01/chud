import { useMutation, type MutationMeta, useQuery, useQueryClient } from '@tanstack/react-query'
import { authFetch } from '@/lib/api-client'
import type { User, UserPayload } from '@/features/users/types'

const usersKey = ['users'] as const

const getUsers = async (): Promise<User[]> => {
  const res = await authFetch('/users')
  return (await res.json()) as User[]
}

const createUser = async (payload: UserPayload): Promise<User> => {
  const res = await authFetch('/users', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return (await res.json()) as User
}

const updateUser = async ({ id, ...payload }: UserPayload & { id: string }): Promise<User> => {
  const res = await authFetch(`/users/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
  return (await res.json()) as User
}

const deleteUser = async (id: string): Promise<void> => {
  await authFetch(`/users/${encodeURIComponent(id)}`, { method: 'DELETE' })
}

export const useUsers = () => useQuery({ queryKey: usersKey, queryFn: getUsers })

const useUsersMutation = <TVariables, TData>(
  mutationFn: (variables: TVariables) => Promise<TData>,
  meta: MutationMeta,
) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn,
    meta,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: usersKey }),
  })
}

export const useCreateUser = () =>
  useUsersMutation(createUser, { success: 'Dodano użytkownika.', inlineError: true })
export const useUpdateUser = () => useUsersMutation(updateUser, { success: 'Zapisano zmiany.', inlineError: true })
export const useDeleteUser = () => useUsersMutation(deleteUser, { success: 'Usunięto użytkownika.' })

// Lookup of every user by id, for showing names and colors next to activity data.
export const useUsersById = (): Map<string, User> => {
  const { data } = useUsers()
  return new Map((data ?? []).map((user) => [user.id, user]))
}
