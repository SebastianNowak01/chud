import type { ComponentProps } from 'react'
import { DrawablyCard } from 'drawably/react'

const PADDING = {
  md: 'p-3.5 sm:p-5',
  lg: 'p-3.5 sm:p-6',
}

export type CardSize = keyof typeof PADDING

type CardProps = ComponentProps<typeof DrawablyCard> & { size?: CardSize; filled?: boolean }

const FILLED = 'before:absolute before:inset-1 before:-z-10 before:rounded-[10px] before:bg-paper'

export function Card({ size = 'md', filled = false, className, ...rest }: CardProps) {
  return <DrawablyCard className={`${PADDING[size]} ${filled ? FILLED : ''} ${className ?? ''}`} {...rest} />
}
