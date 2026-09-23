import type { ReactNode } from 'react'

interface FieldProps {
  label: string
  htmlFor?: string
  className?: string
  children: ReactNode
}

export function Field({ label, htmlFor, className, children }: FieldProps) {
  const Label = htmlFor ? 'label' : 'span'
  return (
    <div className={`flex flex-col gap-1 [&_.drawably-input]:w-full [&_input]:w-full ${className ?? ''}`}>
      <Label htmlFor={htmlFor} className="text-[14px] text-muted">
        {label}
      </Label>
      {children}
    </div>
  )
}
