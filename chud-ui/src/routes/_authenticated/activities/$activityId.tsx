import { createFileRoute } from '@tanstack/react-router'
import { ActivityPage } from '@/features/activities/pages/ActivityPage'

export const Route = createFileRoute('/_authenticated/activities/$activityId')({
  component: ActivityPage,
})
