import { useId, useMemo, type CSSProperties } from 'react'
import { roughCircle } from 'drawably'
import face0 from '@/assets/moods/face-0.png'
import face1 from '@/assets/moods/face-1.png'
import face2 from '@/assets/moods/face-2.png'
import face3 from '@/assets/moods/face-3.png'
import face4 from '@/assets/moods/face-4.png'
import face5 from '@/assets/moods/face-5.png'
import face6 from '@/assets/moods/face-6.png'
import face7 from '@/assets/moods/face-7.png'
import face8 from '@/assets/moods/face-8.png'
import { weekFace } from '@/features/stats/mood'

const FACES = [face0, face1, face2, face3, face4, face5, face6, face7, face8]

interface MoodWojakProps {
  score: number
  seed?: number
  className?: string
  style?: CSSProperties
}

export function MoodWojak({ score, seed = 11, className, style }: MoodWojakProps) {
  const clipId = useId()
  const outline = useMemo(() => roughCircle(50, 50, 44, { seed, roughness: 1 }), [seed])

  return (
    <svg viewBox="0 0 100 100" className={className} style={style} aria-hidden="true">
      <clipPath id={clipId}>
        <circle cx="50" cy="50" r="44" />
      </clipPath>
      <circle cx="50" cy="50" r="44" className="fill-white" />
      <image href={FACES[weekFace(score)]} x="6" y="6" width="88" height="88" clipPath={`url(#${clipId})`} />
      <path d={outline} className="fill-none stroke-ink [stroke-width:1.8]" />
    </svg>
  )
}
