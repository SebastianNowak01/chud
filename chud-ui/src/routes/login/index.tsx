import { createFileRoute, redirect } from '@tanstack/react-router'
import { LoginPage } from '@/features/auth/pages/LoginPage'
import { readSession } from '@/features/auth/lib/token'

export const Route = createFileRoute('/login/')({
  beforeLoad: () => {
    if (readSession()) {
      throw redirect({ to: '/' })
    }
  },
  component: LoginPage,
})
