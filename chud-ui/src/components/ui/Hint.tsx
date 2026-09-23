import type { ReactNode } from 'react'

interface HintProps {
  as?: 'p' | 'span'
  className?: string
  children: ReactNode
}

export function Hint({ as: Tag = 'p', className, children }: HintProps) {
  return <Tag className={`text-[13px] text-muted ${className ?? ''}`}>{children}</Tag>
}
