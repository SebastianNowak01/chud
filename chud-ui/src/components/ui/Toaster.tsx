import { useEffect, useLayoutEffect, useRef } from 'react'
import { Card } from '@/components/ui/Card'
import { toast, useToasts, type Toast } from '@/lib/toast'

const DURATION = 6000

const TONE = {
  error: '[--drawably-stroke:var(--color-danger)]',
  success: '[--drawably-stroke:var(--color-done)]',
}

export function Toaster() {
  const toasts = useToasts()
  const layer = useRef<HTMLDivElement>(null)

  useLayoutEffect(() => {
    const el = layer.current
    if (!el) return
    if (el.matches(':popover-open')) el.hidePopover()
    if (toasts.length > 0) el.showPopover()
  }, [toasts])

  useEffect(() => {
    const offline = () => toast.error('Jesteś offline. Sprawdź połączenie z internetem.', { id: 'offline', sticky: true })
    const online = () => {
      toast.dismiss('offline')
      toast.success('Znowu jesteś online.')
    }
    window.addEventListener('offline', offline)
    window.addEventListener('online', online)
    return () => {
      window.removeEventListener('offline', offline)
      window.removeEventListener('online', online)
    }
  }, [])

  return (
    <div
      ref={layer}
      popover="manual"
      aria-label="Powiadomienia"
      className="inset-auto right-4 bottom-4 left-4 m-0 flex w-auto flex-col gap-2 overflow-visible border-0 bg-transparent p-0 text-ink sm:left-auto sm:w-[360px]"
    >
      {toasts.map((t) => (
        <ToastItem key={t.id} toast={t} />
      ))}
    </div>
  )
}

function ToastItem({ toast: t }: { toast: Toast }) {
  const timer = useRef<number | undefined>(undefined)

  const start = () => {
    if (t.sticky) return
    window.clearTimeout(timer.current)
    timer.current = window.setTimeout(() => toast.dismiss(t.id), DURATION)
  }
  const pause = () => window.clearTimeout(timer.current)

  useEffect(() => {
    if (t.sticky) return
    timer.current = window.setTimeout(() => toast.dismiss(t.id), DURATION)
    return () => window.clearTimeout(timer.current)
  }, [t.id, t.sticky, t.version])

  return (
    <div role={t.kind === 'error' ? 'alert' : 'status'} onPointerEnter={pause} onPointerLeave={start}>
      <Card filled className={`flex items-start gap-3 ${TONE[t.kind]}`}>
        <p className="m-0 flex-1 text-[15px]">{t.message}</p>
        <button
          type="button"
          aria-label="Zamknij"
          className="-m-1 cursor-pointer rounded bg-transparent p-1 text-[18px] leading-none text-muted hover:text-ink"
          onClick={() => toast.dismiss(t.id)}
        >
          ×
        </button>
      </Card>
    </div>
  )
}
