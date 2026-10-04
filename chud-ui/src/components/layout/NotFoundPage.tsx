import { Link } from '@tanstack/react-router'
import { usePageTitle } from '@/lib/use-page-title'

export function NotFoundPage() {
  usePageTitle('Nie znaleziono')
  return (
    <div className="mx-auto flex max-w-[1120px] flex-col gap-2 px-3 pt-4 pb-12 sm:gap-3 sm:px-4 sm:pt-6 sm:pb-16">
      <h1 className="text-[24px] sm:text-[2em]">Nie znaleziono</h1>
      <Link to="/">Strona główna</Link>
    </div>
  )
}
