import { createFileRoute } from '@tanstack/react-router'
import { DashboardPage } from '@/features/dashboard/pages/DashboardPage'

interface DashboardSearch {
  osoby?: string
  aktywnosci?: string
}

const optionalString = (value: unknown) => (typeof value === 'string' && value ? value : undefined)

export const Route = createFileRoute('/_authenticated/')({
  validateSearch: (search: Record<string, unknown>): DashboardSearch => ({
    osoby: optionalString(search.osoby),
    aktywnosci: optionalString(search.aktywnosci),
  }),
  component: DashboardPage,
})
