import { describe, expect, it } from 'vitest'
import { ApiError, errorMessage, NETWORK_ERROR, SERVER_ERROR, statusMessage, UNEXPECTED_ERROR } from '@/lib/api-client'

describe('statusMessage', () => {
  it('uses the server message for client errors, capitalized', () => {
    expect(statusMessage(409, 'nazwa jest już zajęta')).toBe('Nazwa jest już zajęta')
  })

  it('falls back to a message per status', () => {
    expect(statusMessage(404)).toBe('Nie znaleziono.')
    expect(statusMessage(429, '  ')).toBe('Za dużo prób. Poczekaj chwilę.')
  })

  it('hides server error details', () => {
    expect(statusMessage(500, 'Internal server error')).toBe(SERVER_ERROR)
    expect(statusMessage(502)).toBe(SERVER_ERROR)
  })

  it('treats no response as a network error', () => {
    expect(statusMessage(0)).toBe(NETWORK_ERROR)
  })

  it('has a generic message for unknown statuses', () => {
    expect(statusMessage(418)).toBe(UNEXPECTED_ERROR)
  })
})

describe('errorMessage', () => {
  it('shows API errors as they are', () => {
    expect(errorMessage(new ApiError(400, 'Zła data'))).toBe('Zła data')
  })

  it('never shows raw exceptions', () => {
    expect(errorMessage(new TypeError('x is undefined'))).toBe(UNEXPECTED_ERROR)
    expect(errorMessage('boom')).toBe(UNEXPECTED_ERROR)
  })
})
