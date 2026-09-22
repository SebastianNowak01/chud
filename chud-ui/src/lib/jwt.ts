export interface JwtClaims {
  exp: number
  username: string
  is_admin: boolean
}

const isJwtClaims = (data: unknown): data is JwtClaims => {
  if (typeof data !== 'object' || data === null) {
    return false
  }
  const obj = data as Record<string, unknown>
  return (
    typeof obj.username === 'string' &&
    typeof obj.exp === 'number' &&
    typeof obj.is_admin === 'boolean'
  )
}

const decodeBase64Url = (value: string): string => {
  const base64 = value.replace(/-/g, '+').replace(/_/g, '/')
  return atob(base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '='))
}

// Returns the token's claims if it is well-formed and not expired. The signature is checked by the server.
export const decodeValidJwt = (token: string): JwtClaims | null => {
  try {
    const claims: unknown = JSON.parse(decodeBase64Url(token.split('.')[1]))
    if (!isJwtClaims(claims) || claims.exp * 1000 < Date.now()) {
      return null
    }
    return claims
  } catch {
    return null
  }
}
