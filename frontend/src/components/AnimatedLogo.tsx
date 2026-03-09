import { useState, useEffect, useRef } from 'react'

const WORDS = ['Sweet', 'Sour', 'Salty', 'Bitter', 'Umami']
const SLIDE_MS = 450
const PAUSE_MS = 1200

export default function AnimatedLogo({ fontSize = 20 }: { fontSize?: number }) {
  const [index, setIndex] = useState(0)
  const [animating, setAnimating] = useState(false)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    const startCycle = () => {
      setAnimating(true)
      timerRef.current = setTimeout(() => {
        setIndex((prev) => (prev + 1) % WORDS.length)
        setAnimating(false)
        timerRef.current = setTimeout(startCycle, PAUSE_MS)
      }, SLIDE_MS)
    }

    timerRef.current = setTimeout(startCycle, PAUSE_MS)
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [])

  const nextIndex = (index + 1) % WORDS.length
  const slotHeight = fontSize * 1.4

  const wordStyle = (word: string): React.CSSProperties => ({
    fontSize: word === 'Salty' ? fontSize : fontSize * 0.78,
    fontFamily: word === 'Salty' ? "'Raleway', sans-serif" : "'Pacifico', cursive",
    fontWeight: word === 'Salty' ? 600 : 400,
    lineHeight: `${slotHeight}px`,
    whiteSpace: 'nowrap',
  })

  return (
    <div style={{ display: 'flex', alignItems: 'center' }}>
      <div
        style={{
          height: slotHeight,
          overflow: 'hidden',
          position: 'relative',
          minWidth: fontSize * 3.2,
          textAlign: 'right',
        }}
      >
        {/* Current word */}
        <div
          style={{
            ...wordStyle(WORDS[index]),
            position: 'absolute',
            right: 0,
            bottom: animating ? slotHeight : 0,
            transition: animating ? `bottom ${SLIDE_MS}ms ease-in-out` : 'none',
          }}
        >
          {WORDS[index]}
        </div>
        {/* Next word */}
        <div
          style={{
            ...wordStyle(WORDS[nextIndex]),
            position: 'absolute',
            right: 0,
            bottom: animating ? 0 : -slotHeight,
            transition: animating ? `bottom ${SLIDE_MS}ms ease-in-out` : 'none',
          }}
        >
          {WORDS[nextIndex]}
        </div>
      </div>
      <span
        style={{
          fontSize,
          fontFamily: "'Raleway', sans-serif",
          fontWeight: 700,
          lineHeight: `${slotHeight}px`,
        }}
      >
        Bytes
      </span>
    </div>
  )
}
