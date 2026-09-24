import { useState, type FormEvent } from 'react'
import { DrawablyButton, DrawablyCheckbox, DrawablyInput } from 'drawably/react'
import { FormError } from '@/components/ui/FormError'
import { Field } from '@/components/ui/Field'
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
    <form className="flex flex-col gap-4" onSubmit={submit}>
      <Field label="Cel" htmlFor="plan-title">
        <DrawablyInput
          id="plan-title"
          placeholder="Siłownia trzy razy w tygodniu"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          maxLength={100}
          required
        />
      </Field>
      <Field label="Dni">
        <div className="flex flex-wrap items-center gap-3">
          {WEEKDAYS.map(({ key, label }) => (
            <label key={key} className="inline-flex min-h-11 cursor-pointer items-center gap-1.5 sm:min-h-0">
              <DrawablyCheckbox checked={days.has(key)} onChange={() => toggleDay(key)} />
              {label}
            </label>
          ))}
        </div>
      </Field>
      <div className="flex flex-wrap items-center gap-3">
        <Field label="Od" htmlFor="plan-starts-on" className="flex-[1_1_140px] sm:flex-initial">
          <DrawablyInput
            id="plan-starts-on"
            type="date"
            value={startsOn}
            onChange={(e) => setStartsOn(e.target.value)}
            required
          />
        </Field>
        <Field label="Do (opcjonalnie)" htmlFor="plan-ends-on" className="flex-[1_1_140px] sm:flex-initial">
          <DrawablyInput
            id="plan-ends-on"
            type="date"
            value={endsOn}
            min={startsOn}
            onChange={(e) => setEndsOn(e.target.value)}
          />
        </Field>
      </div>
      <FormError error={createPlan.error} />
      <div className="flex flex-wrap items-center gap-3">
        <DrawablyButton type="submit" variant="solid" state={buttonState(createPlan.status)}>
          Utwórz
        </DrawablyButton>
        <DrawablyButton type="button" tone="neutral" onClick={onDone}>
          Anuluj
        </DrawablyButton>
      </div>
    </form>
  )
}
