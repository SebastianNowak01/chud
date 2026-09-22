import { useState } from 'react'

// useState that remembers its value in localStorage; storage may be unavailable, so failures are ignored.
export const useStoredState = <T extends string>(key: string, initial: T, allowed: readonly T[]) => {
  const [value, setValue] = useState<T>(() => {
    try {
      const stored = localStorage.getItem(key) as T | null
      return stored && allowed.includes(stored) ? stored : initial
    } catch {
      return initial
    }
  })

  const update = (next: T) => {
    setValue(next)
    try {
      localStorage.setItem(key, next)
    } catch {
      // Not persisted; the choice still applies for this visit.
    }
  }

  return [value, update] as const
}
