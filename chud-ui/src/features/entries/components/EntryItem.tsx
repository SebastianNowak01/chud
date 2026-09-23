import { DrawablyBadge } from 'drawably/react'
import { UserTag } from '@/components/common/UserTag'
import { Hint } from '@/components/ui/Hint'
import { mediaUrl, useEntryMedia } from '@/features/entries/entries-api'
import type { Entry } from '@/features/entries/types'
import type { User } from '@/features/users/types'
import { apiDate, formatDate, formatDateTime } from '@/lib/dates'

const MEDIA = 'block w-full max-w-full rounded-lg sm:max-h-[200px] sm:w-auto'

export function EntryItem({ entry, user }: { entry: Entry; user: User | undefined }) {
  return (
    <li className="flex flex-col gap-2 border-b border-dashed border-rule py-3.5">
      <div className="flex flex-wrap items-center gap-3">
        <UserTag user={user} />
        <Hint as="span">{formatDateTime(entry.occurredAt)}</Hint>
        {entry.scheduledFor && (
          <DrawablyBadge>
            {entry.excused ? 'excused' : 'planned'} {formatDate(apiDate(entry.scheduledFor))}
          </DrawablyBadge>
        )}
      </div>
      {entry.description && <p className="m-0 whitespace-pre-wrap">{entry.description}</p>}
      <EntryMedia entryId={entry.id} />
    </li>
  )
}

function EntryMedia({ entryId }: { entryId: string }) {
  const media = useEntryMedia(entryId)
  if (!media.data?.length) {
    return null
  }

  return (
    <div className="flex flex-wrap gap-2">
      {media.data.map((m) =>
        m.contentType.startsWith('video/') ? (
          <video key={m.id} src={mediaUrl(m.id)} controls preload="metadata" className={MEDIA} />
        ) : (
          <a key={m.id} href={mediaUrl(m.id)} target="_blank" rel="noreferrer">
            <img src={mediaUrl(m.id)} alt="" loading="lazy" className={MEDIA} />
          </a>
        ),
      )}
    </div>
  )
}
