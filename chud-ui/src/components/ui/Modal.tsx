import { useEffect, useId, useRef, type ReactNode } from 'react'
import { Card } from '@/components/ui/Card'

interface ModalProps {
  title: string
  onClose: () => void
  children: ReactNode
}

export function Modal({ title, onClose, children }: ModalProps) {
  const dialog = useRef<HTMLDialogElement>(null)
  const pressedBackdrop = useRef(false)
  const titleId = useId()

  useEffect(() => {
    const opener = document.activeElement
    dialog.current?.showModal()
    return () => {
      if (opener instanceof HTMLElement) opener.focus()
    }
  }, [])

  return (
    <dialog
      ref={dialog}
      aria-labelledby={titleId}
      className="m-auto max-h-[calc(100dvh-32px)] w-[calc(100%-32px)] max-w-[480px] overflow-y-auto bg-transparent p-1 text-ink backdrop:bg-ink/40"
      onCancel={(e) => {
        e.preventDefault()
        onClose()
      }}
      onPointerDown={(e) => {
        pressedBackdrop.current = e.target === e.currentTarget
      }}
      onClick={(e) => {
        if (pressedBackdrop.current && e.target === e.currentTarget) onClose()
      }}
    >
      <Card filled className="flex flex-col gap-4">
        <h2 id={titleId}>{title}</h2>
        {children}
      </Card>
    </dialog>
  )
}
