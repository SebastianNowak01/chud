import { useEffect } from 'react'
import { Link, useRouter, type ErrorComponentProps } from '@tanstack/react-router'
import { DrawablyButton } from 'drawably/react'
import { Card } from '@/components/ui/Card'
import { Hint } from '@/components/ui/Hint'
import { errorMessage } from '@/lib/api-client'
import { usePageTitle } from '@/lib/use-page-title'

export function ErrorPage({ error, reset }: ErrorComponentProps) {
  usePageTitle('Błąd')
  const router = useRouter()

  useEffect(() => {
    console.error(error)
  }, [error])

  const retry = () => {
    reset()
    void router.invalidate()
  }

  return (
    <div className="grid min-h-[50svh] place-items-center p-4">
      <Card className="flex w-full max-w-[420px] flex-col gap-3">
        <h2>Coś się wysypało</h2>
        <Hint>{errorMessage(error)}</Hint>
        <div className="flex flex-wrap items-center gap-3">
          <DrawablyButton variant="solid" onClick={retry}>
            Spróbuj ponownie
          </DrawablyButton>
          <Link to="/">Strona główna</Link>
        </div>
      </Card>
    </div>
  )
}
