import { DrawablyButton } from 'drawably/react'
import { Hint } from '@/components/ui/Hint'
import { ApiError } from '@/lib/api-client'

const isMissing = (error: unknown) => error instanceof ApiError && error.status === 404

export function LoadError({ error, onRetry }: { error: unknown; onRetry?: () => void }) {
  if (!error) return null
  const missing = isMissing(error)
  return (
    <div className="flex flex-wrap items-center gap-3">
      <Hint as="span">{missing ? 'Nie znaleziono.' : 'Nie udało się wczytać danych.'}</Hint>
      {!missing && onRetry && (
        <DrawablyButton tone="neutral" onClick={onRetry}>
          Spróbuj ponownie
        </DrawablyButton>
      )}
    </div>
  )
}
