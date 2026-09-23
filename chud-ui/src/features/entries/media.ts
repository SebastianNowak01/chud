import { MAX_FILE_COUNT, MAX_FILE_SIZE, MEDIA_TYPES } from '@/features/entries/types'

export const isMediaType = (contentType: string) => MEDIA_TYPES.includes(contentType)

export const mediaFilesError = (files: File[]): string | null => {
  if (files.length > MAX_FILE_COUNT) {
    return `Maksymalnie ${MAX_FILE_COUNT} plików na wpis.`
  }
  const wrongType = files.find((f) => !isMediaType(f.type))
  if (wrongType) {
    return `Plik ${wrongType.name}: można dołączać tylko zdjęcia (JPG, PNG, GIF, WebP) i filmy (MP4, WebM, MOV).`
  }
  const tooBig = files.find((f) => f.size > MAX_FILE_SIZE)
  if (tooBig) {
    return `Plik ${tooBig.name} ma więcej niż 10 MB.`
  }
  return null
}
