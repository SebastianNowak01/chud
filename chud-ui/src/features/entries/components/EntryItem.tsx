import { DrawablyBadge } from 'drawably/react'
import { UserTag } from '@/components/common/UserTag'
import { mediaUrl, useEntryMedia } from '@/features/entries/entries-api'
import type { Entry } from '@/features/entries/types'
import type { User } from '@/features/users/types'
import { apiDate, formatDate, formatDateTime } from '@/lib/dates'

export function EntryItem({ entry, user }: { entry: Entry; user: User | undefined }) {
  return (
    <li className="entry">
      <div className="row">
        <UserTag user={user} />
        <span className="hint">{formatDateTime(entry.occurredAt)}</span>
        {entry.scheduledFor && (
          <DrawablyBadge>
            {entry.excused ? 'excused' : 'planned'} {formatDate(apiDate(entry.scheduledFor))}
          </DrawablyBadge>
        )}
      </div>
      {entry.description && <p className="entry__description">{entry.description}</p>}
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
    <div className="media">
      {media.data.map((m) =>
        m.contentType.startsWith('video/') ? (
          <video key={m.id} src={mediaUrl(m.id)} controls preload="metadata" />
        ) : (
          <a key={m.id} href={mediaUrl(m.id)} target="_blank" rel="noreferrer">
            <img src={mediaUrl(m.id)} alt="" loading="lazy" />
          </a>
        ),
      )}
    </div>
  )
}
