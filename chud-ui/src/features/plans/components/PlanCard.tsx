import { useState } from 'react'
import { DrawablyButton } from 'drawably/react'
import { UserTag } from '@/components/common/UserTag'
import { Badge } from '@/components/ui/Badge'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { Hint } from '@/components/ui/Hint'
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

const STATUS_STROKE: Record<OccurrenceStatus, string> = {
  done: '[--drawably-stroke:var(--color-done)]',
  excused: '[--drawably-stroke:var(--color-excused)]',
  missed: '[--drawably-stroke:var(--color-danger)]',
  todo: '',
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
    <Card className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h3>{plan.title}</h3>
          <div className="flex flex-wrap items-center gap-3">
            <UserTag user={owner} />
            <Hint as="span">{days.join(', ')}</Hint>
          </div>
        </div>
        {isMine &&
          (confirmingDelete ? (
            <div className="flex flex-wrap items-center gap-3">
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

      {occurrences.length === 0 && <Hint>No planned days in the last two weeks or the next week.</Hint>}
      <ul className="m-0 flex list-none flex-col p-0">
        {occurrences.map((o) => (
          <li key={o.date} className={`flex flex-wrap items-center gap-3 py-1.5 ${STATUS_STROKE[o.status]}`}>
            <span className="min-w-[110px]">{formatDate(o.date)}</span>
            <Badge>{STATUS_LABEL[o.status]}</Badge>
            {o.entry?.description && <Hint as="span">{o.entry.description}</Hint>}
            {isMine && !o.entry && (
              <span className="flex flex-wrap items-center gap-3 w-full sm:ml-auto sm:w-auto">
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
      {deletePlan.error && <ErrorText>{deletePlan.error.message}</ErrorText>}
    </Card>
  )
}
