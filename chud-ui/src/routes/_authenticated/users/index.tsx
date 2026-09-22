import { createFileRoute, redirect } from '@tanstack/react-router'
import { UsersPage } from '@/features/users/pages/UsersPage'

export const Route = createFileRoute('/_authenticated/users/')({
  beforeLoad: ({ context }) => {
    if (!context.session.is_admin) {
      throw redirect({ to: '/' })
    }
  },
  component: UsersPage,
})
