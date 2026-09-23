import { useId, useMemo } from 'react'
import { roughCircle, scribbleFill } from 'drawably'

// A round swatch scribbled in with a pen, matching drawably's sketches.
export function ColorSwatch({ color, seed = 7 }: { color: string; seed?: number }) {
  const clipId = useId()
  const { fill, outline } = useMemo(
    () => ({
      fill: scribbleFill(4, 4, 92, 92, { seed, roughness: 1.2 }),
      outline: roughCircle(50, 50, 44, { seed, roughness: 1 }),
    }),
    [seed],
  )

  return (
    <svg viewBox="0 0 100 100" className="color-swatch" aria-hidden="true">
      <clipPath id={clipId}>
        <circle cx="50" cy="50" r="44" />
      </clipPath>
      <path
        d={fill}
        clipPath={`url(#${clipId})`}
        fill="none"
        stroke={color}
        strokeWidth="3.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path d={outline} className="color-swatch__outline" />
    </svg>
  )
}
