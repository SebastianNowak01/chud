import { describe, expect, it } from 'vitest'
import { mediaFilesError } from '@/features/entries/media'
import { MAX_FILE_COUNT, MAX_FILE_SIZE } from '@/features/entries/types'

const file = (name: string, type: string, size = 10) => new File([new Uint8Array(size)], name, { type })

describe('mediaFilesError', () => {
  it('accepts photos and videos', () => {
    expect(mediaFilesError([file('a.jpg', 'image/jpeg'), file('b.mov', 'video/quicktime')])).toBeNull()
  })

  it('rejects other types', () => {
    expect(mediaFilesError([file('x.svg', 'image/svg+xml')])).toContain('x.svg')
    expect(mediaFilesError([file('x.html', 'text/html')])).toContain('x.html')
    expect(mediaFilesError([file('x', '')])).not.toBeNull()
  })

  it('rejects too big files', () => {
    expect(mediaFilesError([file('big.png', 'image/png', MAX_FILE_SIZE + 1)])).toContain('10 MB')
  })

  it('rejects too many files', () => {
    const files = Array.from({ length: MAX_FILE_COUNT + 1 }, (_, i) => file(`${i}.png`, 'image/png'))
    expect(mediaFilesError(files)).toContain(`${MAX_FILE_COUNT}`)
  })
})
