import { apiFetch } from '@/lib/api-client'
import type { User } from '@/features/users/types'

export interface LoginPayload {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}

const login = async (payload: LoginPayload): Promise<LoginResponse> => {
  const res = await apiFetch('/auth/login', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return (await res.json()) as LoginResponse
}

export const AuthApi = {
  login,
}
