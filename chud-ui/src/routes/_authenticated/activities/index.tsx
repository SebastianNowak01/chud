import { createFileRoute } from '@tanstack/react-router'
import { ActivitiesPage } from '@/features/activities/pages/ActivitiesPage'

export const Route = createFileRoute('/_authenticated/activities/')({
  component: ActivitiesPage,
})
