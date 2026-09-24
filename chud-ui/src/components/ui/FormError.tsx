import { errorMessage } from '@/lib/api-client'

export function FormError({ error }: { error: unknown }) {
  if (!error) return null
  return (
    <p role="alert" className="text-[14px] text-danger">
      {errorMessage(error)}
    </p>
  )
}
