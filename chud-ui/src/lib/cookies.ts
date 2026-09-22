export const setCookie = (
  name: string,
  value: string,
  options?: {
    maxAge?: number
    path?: string
    sameSite?: string
  },
): void => {
  let cookieString = `${encodeURIComponent(name)}=${encodeURIComponent(value)}`

  if (options?.maxAge !== undefined) {
    cookieString += `; Max-Age=${options.maxAge}`
  }

  if (options?.path) {
    cookieString += `; Path=${options.path}`
  }

  if (options?.sameSite) {
    cookieString += `; SameSite=${options.sameSite}`
  }

  document.cookie = cookieString
}

export const getCookie = (name: string): string | null => {
  for (const cookie of document.cookie.split('; ')) {
    const [cookieName, cookieValue] = cookie.split('=')
    if (decodeURIComponent(cookieName) === name) {
      return decodeURIComponent(cookieValue)
    }
  }
  return null
}

export const deleteCookie = (name: string): void => {
  setCookie(name, '', { maxAge: -1, path: '/' })
}
