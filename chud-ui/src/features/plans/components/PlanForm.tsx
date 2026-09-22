import { useState, type FormEvent } from 'react'
import { DrawablyButton, DrawablyCard, DrawablyCheckbox, DrawablyInput } from 'drawably/react'
import { useCreatePlan } from '@/features/plans/plans-api'
import { WEEKDAYS, type PlanPayload, type Weekday } from '@/features/plans/types'
import { buttonState } from '@/lib/button-state'
import { toDateString } from '@/lib/dates'

export function PlanForm({ activityId, onDone }: { activityId: string; onDone: () => void }) {
  const [title, setTitle] = useState('')
  const [days, setDays] = useState<Set<Weekday>>(new Set())
  const [startsOn, setStartsOn] = useState(toDateString(new Date()))
  const [endsOn, setEndsOn] = useState('')
  const createPlan = useCreatePlan(activityId)

  const toggleDay = (day: Weekday) => {
    const next = new Set(days)
    if (next.has(day)) {
      next.delete(day)
    } else {
      next.add(day)
    }
    setDays(next)
  }

  const submit = (e: FormEvent) => {
    e.preventDefault()
    const payload: PlanPayload = {
      title,
      monday: days.has('monday'),
      tuesday: days.has('tuesday'),
      wednesday: days.has('wednesday'),
      thursday: days.has('thursday'),
      friday: days.has('friday'),
      saturday: days.has('saturday'),
      sunday: days.has('sunday'),
      startsOn,
      endsOn: endsOn || null,
    }
    createPlan.mutate(payload, { onSuccess: onDone })
  }

  return (
    <DrawablyCard className="card">
      <form className="stack" onSubmit={submit}>
        <h2>New plan</h2>
        <div className="field">
          <label htmlFor="plan-title">Goal</label>
          <DrawablyInput
            id="plan-title"
            placeholder="Gym three times a week"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
          />
        </div>
        <div className="field">
          <span className="label">Days</span>
          <div className="row">
            {WEEKDAYS.map(({ key, label }) => (
              <label key={key} className="checkbox">
                <DrawablyCheckbox checked={days.has(key)} onChange={() => toggleDay(key)} />
                {label}
              </label>
            ))}
          </div>
        </div>
        <div className="row">
          <div className="field">
            <label htmlFor="plan-starts-on">From</label>
            <DrawablyInput
              id="plan-starts-on"
              type="date"
              value={startsOn}
              onChange={(e) => setStartsOn(e.target.value)}
              required
            />
          </div>
          <div className="field">
            <label htmlFor="plan-ends-on">Until (optional)</label>
            <DrawablyInput
              id="plan-ends-on"
              type="date"
              value={endsOn}
              min={startsOn}
              onChange={(e) => setEndsOn(e.target.value)}
            />
          </div>
        </div>
        {createPlan.error && <p className="error">{createPlan.error.message}</p>}
        <div className="row">
          <DrawablyButton type="submit" variant="solid" state={buttonState(createPlan.status)}>
            Create
          </DrawablyButton>
          <DrawablyButton type="button" tone="neutral" onClick={onDone}>
            Cancel
          </DrawablyButton>
        </div>
      </form>
    </DrawablyCard>
  )
}
