import { useState } from 'react'
import { getRouteApi, Link } from '@tanstack/react-router'
import { DrawablyButton } from 'drawably/react'
import { UserTag } from '@/components/common/UserTag'
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
      <div className="stack">
        <p className="error">{activity.error.message}</p>
        <Link to="/activities">Back to activities</Link>
      </div>
    )
  }

  const openEntryForm = (plannedDay?: PlannedDay) => {
    setForm({ kind: 'entry', plannedDay })
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  return (
    <div className="stack">
      <Link to="/activities" className="hint">
        ← All activities
      </Link>
      <div className="header">
        <div>
          <h2>{activity.data?.name ?? '…'}</h2>
          {activity.data?.description && <p className="hint">{activity.data.description}</p>}
        </div>
        {form.kind === 'none' && (
          <div className="row">
            <DrawablyButton variant="solid" onClick={() => openEntryForm()}>
              Log entry
            </DrawablyButton>
            <DrawablyButton onClick={() => setForm({ kind: 'plan' })}>New plan</DrawablyButton>
          </div>
        )}
      </div>

      {members.data && members.data.length > 0 && (
        <div className="row">
          <span className="hint">Members:</span>
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
        <section className="stack">
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

      <section className="stack">
        <h3>Latest entries</h3>
        {entries.isPending && <p className="hint">Loading…</p>}
        {entries.error && <p className="error">{entries.error.message}</p>}
        {entries.data?.length === 0 && <p className="hint">Nothing logged yet.</p>}
        <ul className="entries">
          {entries.data?.map((entry) => (
            <EntryItem key={entry.id} entry={entry} user={usersById.get(entry.userId)} />
          ))}
        </ul>
      </section>
    </div>
  )
}
