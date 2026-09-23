import { useState } from 'react'
import { getRouteApi, Link } from '@tanstack/react-router'
import { DrawablyButton } from 'drawably/react'
import { UserTag } from '@/components/common/UserTag'
import { ErrorText } from '@/components/ui/ErrorText'
import { Hint } from '@/components/ui/Hint'
import { useActivity, useMembers } from '@/features/activities/activities-api'
import { EntryForm, type PlannedDay } from '@/features/entries/components/EntryForm'
import { EntryItem } from '@/features/entries/components/EntryItem'
import { useEntries } from '@/features/entries/entries-api'
import { PlanCard } from '@/features/plans/components/PlanCard'
import { PlanForm } from '@/features/plans/components/PlanForm'
import { usePlans } from '@/features/plans/plans-api'
import { useUsersById } from '@/features/users/users-api'

const activityRoute = getRouteApi('/_authenticated/activities/$activityId')

type FormState = { kind: 'none' } | { kind: 'entry'; plannedDay?: PlannedDay } | { kind: 'plan' }

export function ActivityPage() {
  const { activityId } = activityRoute.useParams()
  const { session } = activityRoute.useRouteContext()

  const activity = useActivity(activityId)
  const members = useMembers(activityId)
  const entries = useEntries(activityId)
  const plans = usePlans(activityId)
  const usersById = useUsersById()
  const [form, setForm] = useState<FormState>({ kind: 'none' })
  const closeForm = () => setForm({ kind: 'none' })

  if (activity.error) {
    return (
      <div className="flex flex-col gap-4">
        <ErrorText>{activity.error.message}</ErrorText>
        <Link to="/activities">Back to activities</Link>
      </div>
    )
  }

  const openEntryForm = (plannedDay?: PlannedDay) => {
    setForm({ kind: 'entry', plannedDay })
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  return (
    <div className="flex flex-col gap-4">
      <Link to="/activities" className="text-[13px] text-muted">
        ← All activities
      </Link>
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h2>{activity.data?.name ?? '…'}</h2>
          {activity.data?.description && <Hint>{activity.data.description}</Hint>}
        </div>
        {form.kind === 'none' && (
          <div className="flex flex-wrap items-center gap-3">
            <DrawablyButton variant="solid" onClick={() => openEntryForm()}>
              Log entry
            </DrawablyButton>
            <DrawablyButton onClick={() => setForm({ kind: 'plan' })}>New plan</DrawablyButton>
          </div>
        )}
      </div>

      {members.data && members.data.length > 0 && (
        <div className="flex flex-wrap items-center gap-3">
          <Hint as="span">Members:</Hint>
          {members.data.map((member) => (
            <UserTag key={member.id} user={member} />
          ))}
        </div>
      )}

      {form.kind === 'entry' && (
        <EntryForm
          key={form.plannedDay ? `${form.plannedDay.planId}-${form.plannedDay.scheduledFor}` : 'new'}
          activityId={activityId}
          plannedDay={form.plannedDay}
          onDone={closeForm}
        />
      )}
      {form.kind === 'plan' && <PlanForm activityId={activityId} onDone={closeForm} />}

      {plans.data && plans.data.length > 0 && (
        <section className="flex flex-col gap-4">
          <h3>Plans</h3>
          {plans.data.map((plan) => (
            <PlanCard
              key={plan.id}
              plan={plan}
              owner={usersById.get(plan.userId)}
              entries={entries.data ?? []}
              isMine={plan.userId === session.user_id}
              onResolve={openEntryForm}
            />
          ))}
        </section>
      )}

      <section className="flex flex-col gap-4">
        <h3>Latest entries</h3>
        {entries.isPending && <Hint>Loading…</Hint>}
        {entries.error && <ErrorText>{entries.error.message}</ErrorText>}
        {entries.data?.length === 0 && <Hint>Nothing logged yet.</Hint>}
        <ul className="m-0 flex list-none flex-col p-0">
          {entries.data?.map((entry) => (
            <EntryItem key={entry.id} entry={entry} user={usersById.get(entry.userId)} />
          ))}
        </ul>
      </section>
    </div>
  )
}
