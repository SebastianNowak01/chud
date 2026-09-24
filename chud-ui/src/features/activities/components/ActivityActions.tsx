import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { DrawablyButton } from 'drawably/react'
import { FormError } from '@/components/ui/FormError'
import { Hint } from '@/components/ui/Hint'
import { Modal } from '@/components/ui/Modal'
import { useArchiveActivity, useDeleteActivity } from '@/features/activities/activities-api'
import type { Activity } from '@/features/activities/types'
import { buttonState } from '@/lib/button-state'

type Confirming = 'none' | 'delete' | 'archive'

export function ActivityActions({ activity }: { activity: Activity }) {
  const navigate = useNavigate()
  const [confirming, setConfirming] = useState<Confirming>('none')
  const deleteActivity = useDeleteActivity()
  const archiveActivity = useArchiveActivity()
  const close = () => {
    setConfirming('none')
    deleteActivity.reset()
    archiveActivity.reset()
  }

  return (
    <>
      {activity.hasHistory ? (
        <DrawablyButton tone="danger" onClick={() => setConfirming('archive')}>
          Archiwizuj
        </DrawablyButton>
      ) : (
        <DrawablyButton tone="danger" onClick={() => setConfirming('delete')}>
          Usuń
        </DrawablyButton>
      )}

      {confirming === 'delete' && (
        <Modal title={`Usunąć „${activity.name}”?`} onClose={close}>
          <Hint>Aktywność nie ma żadnych planów ani wpisów, więc zniknie bez śladu.</Hint>
          <FormError error={deleteActivity.error} />
          <div className="flex flex-wrap items-center gap-3">
            <DrawablyButton
              tone="danger"
              variant="solid"
              state={buttonState(deleteActivity.status)}
              onClick={() =>
                deleteActivity.mutate(activity.id, { onSuccess: () => void navigate({ to: '/activities' }) })
              }
            >
              Usuń
            </DrawablyButton>
            <DrawablyButton tone="neutral" onClick={close}>
              Anuluj
            </DrawablyButton>
          </div>
        </Modal>
      )}

      {confirming === 'archive' && (
        <Modal title={`Zarchiwizować „${activity.name}”?`} onClose={close}>
          <ul className="m-0 flex list-disc flex-col gap-1 pl-5 text-[15px]">
            <li>Wpisy i zdobyte punkty zostają, historia nadal jest widoczna.</li>
            <li>Trwające plany zostaną zakończone, a te, które się jeszcze nie zaczęły, usunięte.</li>
            <li>Nie da się dodawać nowych wpisów ani planów, dopóki aktywności nie przywrócisz.</li>
          </ul>
          <FormError error={archiveActivity.error} />
          <div className="flex flex-wrap items-center gap-3">
            <DrawablyButton
              tone="danger"
              variant="solid"
              state={buttonState(archiveActivity.status)}
              onClick={() => archiveActivity.mutate(activity.id, { onSuccess: close })}
            >
              Archiwizuj
            </DrawablyButton>
            <DrawablyButton tone="neutral" onClick={close}>
              Anuluj
            </DrawablyButton>
          </div>
        </Modal>
      )}
    </>
  )
}
