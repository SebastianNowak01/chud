import type { ComponentProps } from 'react'
import { DrawablyBadge } from 'drawably/react'

const TONE = {
  inherit: '',
  done: '[--drawably-stroke:var(--color-done)]',
  excused: '[--drawably-stroke:var(--color-excused)]',
  accent: '[--drawably-stroke:var(--color-accent)]',
  danger: '[--drawably-stroke:var(--color-danger)]',
}

export type BadgeTone = keyof typeof TONE

type BadgeProps = ComponentProps<typeof DrawablyBadge> & { tone?: BadgeTone }

export function Badge({ tone = 'inherit', className, ...rest }: BadgeProps) {
  return (
    <DrawablyBadge
      width={1.5}
      className={`px-3 py-1 text-[14px] leading-tight before:absolute before:inset-0.5 before:-z-10 before:rounded-md before:bg-[color-mix(in_srgb,var(--drawably-stroke)_10%,transparent)] ${TONE[tone]} ${className ?? ''}`}
      {...rest}
    />
  )
}
