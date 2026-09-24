import { useSyncExternalStore } from 'react'

export type ToastKind = 'error' | 'success'

export interface Toast {
  id: string
  kind: ToastKind
  message: string
  sticky: boolean
  version: number
}

interface ToastOptions {
  id?: string
  sticky?: boolean
}

export const MAX_TOASTS = 3

let toasts: Toast[] = []
let version = 0
const listeners = new Set<() => void>()

const emit = (next: Toast[]) => {
  toasts = next
  listeners.forEach((listener) => listener())
}

const push = (kind: ToastKind, message: string, options: ToastOptions = {}) => {
  const id = options.id ?? `${kind}:${message}`
  version += 1
  const toast: Toast = { id, kind, message, sticky: options.sticky ?? false, version }
  emit([...toasts.filter((t) => t.id !== id), toast].slice(-MAX_TOASTS))
}

export const toast = {
  error: (message: string, options?: ToastOptions) => push('error', message, options),
  success: (message: string, options?: ToastOptions) => push('success', message, options),
  dismiss: (id: string) => emit(toasts.filter((t) => t.id !== id)),
  clear: () => emit([]),
}

const subscribe = (listener: () => void) => {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

export const currentToasts = () => toasts

export const useToasts = () => useSyncExternalStore(subscribe, currentToasts)
