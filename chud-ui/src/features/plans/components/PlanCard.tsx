import { useState } from 'react'
import { DrawablyButton } from 'drawably/react'
import { UserTag } from '@/components/common/UserTag'
import { Badge } from '@/components/ui/Badge'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { Hint } from '@/components/ui/Hint'
import type { PlannedDay } from '@/features/entries/components/EntryForm'
import type { Entry } from '@/features/entries/types'
import { useDeletePlan, useUpdatePlan } from '@/features/plans/plans-api'
import { WEEKDAYS, type Plan } from '@/features/plans/types'
import { useOccurrences } from '@/features/stats/stats-api'
import { STATUS_LABEL, STATUS_TONE } from '@/features/stats/status'
import type { User } from '@/features/users/types'
import { buttonState } from '@/lib/button-state'
import { addDays, formatDate, toDateString } from '@/lib/dates'

const PAST_DAYS = 14
const FUTURE_DAYS = 7

interface PlanCardProps {
  plan: Plan
  owner: User | undefined
  entries: Entry[]
  isMine: boolean
  onResolve: (day: PlannedDay) => void
}

export function PlanCard({ plan, owner, entries, isMine, onResolve }: PlanCardProps) {
  const [confirming, setConfirming] = useState(false)
  const deletePlan = useDeletePlan(plan.activityId)
  const updatePlan = useUpdatePlan(plan.activityId)

  const now = new Date()
  const today = toDateString(now)
  const occurrences = useOccurrences(toDateString(addDays(now, -PAST_DAYS)), toDateString(addDays(now, FUTURE_DAYS)))
  const planDays = (occurrences.data ?? []).filter((o) => o.planId === plan.id)
  const days = WEEKDAYS.filter(({ key }) => plan[key]).map(({ label }) => label)

  const { startsOn, endsOn } = plan
  const notStarted = startsOn >= today
  const ended = endsOn !== null && endsOn < today
  const todayResolved = planDays.some((o) => o.date === today && o.entryId !== null)
  const endDate = todayResolved ? today : toDateString(addDays(now, -1))

  const confirm = () => {
    if (notStarted) {
      deletePlan.mutate(plan.id)
    } else {
      updatePlan.mutate({ planId: plan.id, payload: { endsOn: endDate } }, { onSuccess: () => setConfirming(false) })
    }
  }
  const mutation = notStarted ? deletePlan : updatePlan

  return (
    <Card className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h3>{plan.title}</h3>
          <div className="flex flex-wrap items-center gap-3">
            <UserTag user={owner} />
            <Hint as="span">{days.join(', ')}</Hint>
            <Hint as="span">
              od {formatDate(startsOn)}
              {endsOn && ` do ${formatDate(endsOn)}`}
            </Hint>
          </div>
        </div>
        {isMine &&
          !ended &&
          (confirming ? (
            <div className="flex flex-wrap items-center gap-3">
              <span>{notStarted ? 'Usunąć plan?' : 'Zakończyć plan?'}</span>
              <DrawablyButton tone="danger" state={buttonState(mutation.status)} onClick={confirm}>
                Tak
              </DrawablyButton>
              <DrawablyButton tone="neutral" onClick={() => setConfirming(false)}>
                Nie
              </DrawablyButton>
            </div>
          ) : (
            <DrawablyButton tone="danger" onClick={() => setConfirming(true)}>
              {notStarted ? 'Usuń' : 'Zakończ plan'}
            </DrawablyButton>
          ))}
      </div>

      {occurrences.error && <ErrorText>{occurrences.error.message}</ErrorText>}
      {occurrences.data && planDays.length === 0 && (
        <Hint>Brak zaplanowanych dni w ostatnich dwóch tygodniach i w najbliższym tygodniu.</Hint>
      )}
      <ul className="m-0 flex list-none flex-col p-0">
        {planDays.map((o) => {
          const description = entries.find((e) => e.id === o.entryId)?.description
          return (
            <li key={o.date} className="flex flex-wrap items-center gap-3 py-1.5">
              <span className="min-w-[110px]">{formatDate(o.date)}</span>
              <Badge tone={STATUS_TONE[o.status]}>{STATUS_LABEL[o.status]}</Badge>
              {description && <Hint as="span">{description}</Hint>}
              {isMine && o.entryId === null && (
                <span className="flex w-full flex-wrap items-center gap-3 sm:ml-auto sm:w-auto">
                  {o.date <= today && (
                    <DrawablyButton onClick={() => onResolve({ planId: plan.id, scheduledFor: o.date, excused: false })}>
                      Zrobione
                    </DrawablyButton>
                  )}
                  <DrawablyButton
                    tone="neutral"
                    onClick={() => onResolve({ planId: plan.id, scheduledFor: o.date, excused: true })}
                  >
                    Wymówka
                  </DrawablyButton>
                </span>
              )}
            </li>
          )
        })}
      </ul>
      {mutation.error && <ErrorText>{mutation.error.message}</ErrorText>}
    </Card>
  )
}
