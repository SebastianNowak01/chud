import { beforeEach, describe, expect, it } from 'vitest'
import { currentToasts, MAX_TOASTS, toast } from '@/lib/toast'

describe('toast', () => {
  beforeEach(() => toast.clear())

  it('shows the same message once and brings it to the front', () => {
    toast.error('A')
    toast.error('B')
    toast.error('A')
    expect(currentToasts().map((t) => t.message)).toEqual(['B', 'A'])
  })

  it(`keeps at most ${MAX_TOASTS}`, () => {
    for (const message of ['1', '2', '3', '4']) toast.success(message)
    expect(currentToasts().map((t) => t.message)).toEqual(['2', '3', '4'])
  })

  it('replaces a toast with the same id', () => {
    toast.error('Offline', { id: 'net', sticky: true })
    toast.error('Nadal offline', { id: 'net', sticky: true })
    expect(currentToasts()).toHaveLength(1)
    expect(currentToasts()[0].message).toBe('Nadal offline')
    toast.dismiss('net')
    expect(currentToasts()).toHaveLength(0)
  })
})
