import type { ReactNode } from 'react'

export function ErrorText({ as: Tag = 'p', children }: { as?: 'p' | 'span'; children: ReactNode }) {
  return <Tag className="text-[14px] text-danger">{children}</Tag>
}
