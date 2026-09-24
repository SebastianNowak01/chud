import { clearToken, readToken } from '@/features/auth/lib/token'

export const API_URL = '/api/v1'

export const NETWORK_ERROR = 'Brak połączenia z serwerem. Sprawdź internet i spróbuj ponownie.'
export const SERVER_ERROR = 'Coś poszło nie tak po stronie serwera. Spróbuj ponownie za chwilę.'
export const UNEXPECTED_ERROR = 'Wystąpił nieoczekiwany błąd. Odśwież stronę i spróbuj ponownie.'

const STATUS_MESSAGE: Record<number, string> = {
  400: 'Nieprawidłowe dane.',
  401: 'Musisz się zalogować.',
  403: 'Nie masz do tego uprawnień.',
  404: 'Nie znaleziono.',
  405: 'Ta operacja nie jest dozwolona.',
  409: 'To koliduje z istniejącymi danymi.',
  413: 'Za duży plik lub żądanie.',
  429: 'Za dużo prób. Poczekaj chwilę.',
}

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

const capitalize = (text: string) => text.charAt(0).toUpperCase() + text.slice(1)

export const statusMessage = (status: number, serverMessage?: string): string => {
  if (status === 0) return NETWORK_ERROR
  if (status >= 500) return SERVER_ERROR
  if (serverMessage?.trim()) return capitalize(serverMessage.trim())
  return STATUS_MESSAGE[status] ?? UNEXPECTED_ERROR
}

export const errorMessage = (error: unknown): string =>
  error instanceof ApiError ? error.message : UNEXPECTED_ERROR

export const isUnauthorized = (error: unknown) => error instanceof ApiError && error.status === 401

const serverMessage = async (res: Response): Promise<string | undefined> => {
  try {
    const body = (await res.json()) as { error?: { message?: string } }
    return body.error?.message
  } catch {
    return undefined
  }
}

export const apiFetch = async (path: string, options?: RequestInit): Promise<Response> => {
  const isJson = typeof options?.body === 'string'
  let res: Response
  try {
    res = await fetch(`${API_URL}${path}`, {
      ...options,
      headers: {
        ...(isJson && { 'Content-Type': 'application/json' }),
        ...options?.headers,
      },
    })
  } catch {
    throw new ApiError(0, NETWORK_ERROR)
  }

  if (!res.ok) {
    throw new ApiError(res.status, statusMessage(res.status, await serverMessage(res)))
  }

  return res
}

let unauthorizedHandler = () => window.location.assign('/login')

export const onUnauthorized = (handler: () => void) => {
  unauthorizedHandler = handler
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
    if (isUnauthorized(error)) {
      clearToken()
      unauthorizedHandler()
    }
    throw error
  }
}
