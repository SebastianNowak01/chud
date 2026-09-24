import { useEffect, useRef, useState } from 'react'
import { Link } from '@tanstack/react-router'
import calmWojak from '@/assets/niemabazyWojak.png'
import hypedWojak from '@/assets/jestbazaWojak.png'
import badToTheBone from '@/assets/badToTheBone.mp3'

// The song is decoded once up front, so playback starts instantly (an <audio> element has to seek and buffer).
let audioContext: AudioContext | null = null
// The mp3 opens with ~0.7 s of silence (measured); playback skips it.
const SONG_START_SECONDS = 0.68
let songPromise: Promise<AudioBuffer> | null = null

const UNLOCK_EVENTS = ['pointerdown', 'keydown', 'touchend'] as const

// Browsers start audio suspended and only let it resume inside a click/tap/keypress handler
// (a hover doesn't count), so the first such gesture anywhere on the page unlocks it.
const unlockAudio = () => {
  if (!audioContext) return
  void audioContext
    .resume()
    .then(() => UNLOCK_EVENTS.forEach((type) => document.removeEventListener(type, unlockAudio, true)))
    .catch(() => {})
}

const loadSong = () => {
  if (!audioContext) {
    audioContext = new AudioContext()
    UNLOCK_EVENTS.forEach((type) => document.addEventListener(type, unlockAudio, true))
  }
  const ctx = audioContext
  songPromise ??= fetch(badToTheBone)
    .then((res) => res.arrayBuffer())
    .then((data) => ctx.decodeAudioData(data))
  return songPromise
}

const IMAGE_SIZE = {
  small: 'size-14 sm:size-[72px]',
  large: 'size-44',
}

interface LogoProps {
  size: 'small' | 'large'
  // Navbar logo links home; the login page one is just a picture.
  linked?: boolean
}

// Wojak that gets hyped and plays "Bad to the Bone" while hovered (ported from SebChan).
export function Logo({ size, linked = false }: LogoProps) {
  const [hovered, setHovered] = useState(false)
  const song = useRef<AudioBuffer | null>(null)
  const playing = useRef<AudioBufferSourceNode | null>(null)

  useEffect(() => {
    loadSong()
      .then((buffer) => (song.current = buffer))
      .catch(() => {})
    return () => playing.current?.stop()
  }, [])

  const start = () => {
    setHovered(true)
    // Silent until the page got its first click/tap/keypress (see unlockAudio); the face still changes.
    if (!audioContext || audioContext.state !== 'running' || !song.current) return
    playing.current?.stop()
    const source = audioContext.createBufferSource()
    source.buffer = song.current
    source.connect(audioContext.destination)
    source.start(0, SONG_START_SECONDS)
    playing.current = source
  }

  const stop = () => {
    setHovered(false)
    playing.current?.stop()
    playing.current = null
  }

  const imageClass = `${IMAGE_SIZE[size]} rounded-full bg-white object-contain object-top p-[4%] [&:not([hidden])]:block`
  // Both faces stay in the DOM, so swapping them never waits for an image to load.
  const image = (
    <span className="block">
      <img src={calmWojak} alt="chud" draggable={false} hidden={hovered} className={imageClass} />
      <img src={hypedWojak} alt="chud" draggable={false} hidden={!hovered} className={imageClass} />
    </span>
  )

  const handlers = { onPointerEnter: start, onPointerLeave: stop }

  return linked ? (
    <Link to="/activities" className="inline-flex leading-[0]" aria-label="chud" {...handlers}>
      {image}
    </Link>
  ) : (
    <span className="inline-flex leading-[0]" {...handlers}>
      {image}
    </span>
  )
}
