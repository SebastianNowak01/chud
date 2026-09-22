import { createFileRoute, redirect } from '@tanstack/react-router'
import { HomePage } from '@/features/home/pages/HomePage'

export const Route = createFileRoute('/_authenticated/')({
  beforeLoad: ({ context }) => {
    if (context.session.is_admin) {
      throw redirect({ to: '/users' })
    }
  },
  component: HomePage,
})
