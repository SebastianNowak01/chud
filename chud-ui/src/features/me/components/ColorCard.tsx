import type { FormEvent } from 'react'
import { DrawablyButton } from 'drawably/react'
import { ColorSwatch } from '@/components/common/ColorSwatch'
import { Card } from '@/components/ui/Card'
import { ErrorText } from '@/components/ui/ErrorText'
import { Hint } from '@/components/ui/Hint'
import { useUpdateMe } from '@/features/me/me-api'
import { buttonState } from '@/lib/button-state'

interface ColorCardProps {
  color: string
  savedColor: string
  onColorChange: (color: string | null) => void
}

export function ColorCard({ color, savedColor, onColorChange }: ColorCardProps) {
  const updateMe = useUpdateMe()
  const changed = color !== savedColor

  const submit = (e: FormEvent) => {
    e.preventDefault()
    updateMe.mutate({ color }, { onSuccess: () => onColorChange(null) })
  }

  return (
    <Card className="flex items-center">
      <form className="flex flex-wrap items-center gap-6" onSubmit={submit}>
        <label className="group flex cursor-pointer flex-col items-center gap-1" title="Wybierz swój kolor">
          <ColorSwatch
            color={color}
            className="size-[104px] rounded-full transition-transform duration-150 ease-[ease] group-hover:scale-[1.04] group-hover:-rotate-6 group-focus-within:outline-2 group-focus-within:outline-offset-4 group-focus-within:outline-ink group-focus-within:outline-dashed"
          />
          <input
            type="color"
            className="sr-only"
            aria-label="Twój kolor"
            value={color}
            onChange={(e) => onColorChange(e.target.value)}
          />
          <Hint as="span">kliknij, aby zmienić</Hint>
        </label>

        <div className="flex min-w-0 flex-col items-start gap-1.5">
          <span className="text-[14px] font-semibold tracking-widest text-muted uppercase">Twój kolor</span>
          <code className="text-[14px] text-muted">{color}</code>
          {updateMe.error && <ErrorText>{updateMe.error.message}</ErrorText>}
          {changed && (
            <div className="flex flex-wrap items-center gap-3">
              <DrawablyButton type="submit" variant="solid" state={buttonState(updateMe.status)}>
                Zapisz kolor
              </DrawablyButton>
              <DrawablyButton type="button" tone="neutral" onClick={() => onColorChange(null)}>
                Cofnij
              </DrawablyButton>
            </div>
          )}
        </div>
      </form>
    </Card>
  )
}
