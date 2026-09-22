import { createFileRoute, redirect } from '@tanstack/react-router'
import { AppLayout } from '@/components/layout/AppLayout'
import { readSession } from '@/features/auth/lib/token'

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: () => {
    const session = readSession()
    if (!session) {
      throw redirect({ to: '/login' })
    }
    return { session }
  },
  component: AppLayout,
})
