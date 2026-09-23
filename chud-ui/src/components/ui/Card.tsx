import type { ComponentProps } from 'react'
import { DrawablyCard } from 'drawably/react'

const PADDING = {
  md: 'p-3.5 sm:p-5',
  lg: 'p-3.5 sm:p-6',
}

export type CardSize = keyof typeof PADDING

type CardProps = ComponentProps<typeof DrawablyCard> & { size?: CardSize }

export function Card({ size = 'md', className, ...rest }: CardProps) {
  return <DrawablyCard className={`${PADDING[size]} ${className ?? ''}`} {...rest} />
}
