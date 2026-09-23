import { createFileRoute } from '@tanstack/react-router'
import { UserProfilePage } from '@/features/users/pages/UserProfilePage'

export const Route = createFileRoute('/_authenticated/users/$userId')({
  component: UserProfilePage,
})
