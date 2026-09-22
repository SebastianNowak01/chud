import { useState } from 'react'
import { DrawablyBadge, DrawablyButton, DrawablyCard } from 'drawably/react'
import { UserTag } from '@/components/common/UserTag'
import type { PlannedDay } from '@/features/entries/components/EntryForm'
import type { Entry } from '@/features/entries/types'
import { planOccurrences, type OccurrenceStatus } from '@/features/plans/occurrences'
import { useDeletePlan } from '@/features/plans/plans-api'
import { WEEKDAYS, type Plan } from '@/features/plans/types'
import type { User } from '@/features/users/types'
import { buttonState } from '@/lib/button-state'
import { addDays, formatDate } from '@/lib/dates'

const PAST_DAYS = 14
const FUTURE_DAYS = 7

const STATUS_LABEL: Record<OccurrenceStatus, string> = {
  done: '✓ done',
  excused: 'excused',
  missed: '✗ missed',
  todo: 'to do',
}

interface PlanCardProps {
  plan: Plan
  owner: User | undefined
  entries: Entry[]
  isMine: boolean
  onResolve: (day: PlannedDay) => void
}

export function PlanCard({ plan, owner, entries, isMine, onResolve }: PlanCardProps) {
  const [confirmingDelete, setConfirmingDelete] = useState(false)
  const deletePlan = useDeletePlan(plan.activityId)

  const today = new Date()
  const occurrences = planOccurrences(plan, entries, addDays(today, -PAST_DAYS), addDays(today, FUTURE_DAYS))
  const days = WEEKDAYS.filter(({ key }) => plan[key]).map(({ label }) => label)

  return (
    <DrawablyCard className="card stack">
      <div className="header">
        <div>
          <h3>{plan.title}</h3>
          <div className="row">
            <UserTag user={owner} />
            <span className="hint">{days.join(', ')}</span>
          </div>
        </div>
        {isMine &&
          (confirmingDelete ? (
            <div className="row">
              <span>Delete plan?</span>
              <DrawablyButton
                tone="danger"
                state={buttonState(deletePlan.status)}
                onClick={() => deletePlan.mutate(plan.id)}
              >
                Yes
              </DrawablyButton>
              <DrawablyButton tone="neutral" onClick={() => setConfirmingDelete(false)}>
                No
              </DrawablyButton>
            </div>
          ) : (
            <DrawablyButton tone="danger" onClick={() => setConfirmingDelete(true)}>
              Delete
            </DrawablyButton>
          ))}
      </div>

      {occurrences.length === 0 && <p className="hint">No planned days in the last two weeks or the next week.</p>}
      <ul className="occurrences">
        {occurrences.map((o) => (
          <li key={o.date} className={`occurrence occurrence--${o.status}`}>
            <span className="occurrence__date">{formatDate(o.date)}</span>
            <DrawablyBadge>{STATUS_LABEL[o.status]}</DrawablyBadge>
            {o.entry?.description && <span className="hint">{o.entry.description}</span>}
            {isMine && !o.entry && (
              <span className="row occurrence__actions">
                <DrawablyButton onClick={() => onResolve({ planId: plan.id, scheduledFor: o.date, excused: false })}>
                  Done
                </DrawablyButton>
                <DrawablyButton
                  tone="neutral"
                  onClick={() => onResolve({ planId: plan.id, scheduledFor: o.date, excused: true })}
                >
                  Excuse
                </DrawablyButton>
              </span>
            )}
          </li>
        ))}
      </ul>
      {deletePlan.error && <p className="error">{deletePlan.error.message}</p>}
    </DrawablyCard>
  )
}
