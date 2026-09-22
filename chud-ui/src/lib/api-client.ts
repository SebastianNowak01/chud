import { clearToken, readToken } from '@/features/auth/lib/token'

export const API_URL = '/api/v1'

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

const errorMessage = async (res: Response): Promise<string> => {
  try {
    const body = (await res.json()) as { error?: { message?: string } }
    return body.error?.message ?? res.statusText
  } catch {
    return res.statusText
  }
}

export const apiFetch = async (path: string, options?: RequestInit): Promise<Response> => {
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!res.ok) {
    throw new ApiError(res.status, await errorMessage(res))
  }

  return res
}

export const authFetch = async (path: string, options?: RequestInit): Promise<Response> => {
  const token = readToken()

  try {
    return await apiFetch(path, {
      ...options,
      headers: {
        ...options?.headers,
        Authorization: `Bearer ${token ?? ''}`,
      },
    })
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      clearToken()
      window.location.assign('/login')
    }
    throw error
  }
}
