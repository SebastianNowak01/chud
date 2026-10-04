import { deleteCookie, getCookie, setCookie } from '@/lib/cookies'
import { decodeValidJwt, type JwtClaims } from '@/lib/jwt'

export const LOGIN_TOKEN_COOKIE = 'LOGIN_TOKEN'

const TOKEN_MAX_AGE_SECONDS = 24 * 60 * 60

export const readToken = (): string | null => getCookie(LOGIN_TOKEN_COOKIE)

export const storeToken = (jwt: string): void => {
  setCookie(LOGIN_TOKEN_COOKIE, jwt, {
    maxAge: TOKEN_MAX_AGE_SECONDS,
    path: '/',
    sameSite: 'Strict',
    secure: window.location.protocol === 'https:',
  })
}

export const clearToken = (): void => {
  deleteCookie(LOGIN_TOKEN_COOKIE)
}

// Claims of the stored token, or null when there is no valid session.
export const readSession = (): JwtClaims | null => {
  const token = readToken()
  return token ? decodeValidJwt(token) : null
}
