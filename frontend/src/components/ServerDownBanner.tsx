import { useState, useEffect, useRef } from 'react'
import { Server, ServerCrash } from 'lucide-react'
import { useNetwork } from '@/contexts/NetworkContext'

export function ServerDownBanner() {
  const { isServerDown, wasServerDown } = useNetwork()
  const [showGreen, setShowGreen] = useState(false)
  const [fadingOut, setFadingOut] = useState(false)
  const fadeTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (wasServerDown && !isServerDown) {
      setShowGreen(true)
      setFadingOut(false)
      fadeTimer.current = setTimeout(() => {
        setFadingOut(true)
        setTimeout(() => setShowGreen(false), 300)
      }, 2200)
    } else if (!wasServerDown) {
      setShowGreen(false)
      setFadingOut(false)
    }
    return () => {
      if (fadeTimer.current) clearTimeout(fadeTimer.current)
    }
  }, [wasServerDown, isServerDown])

  if (isServerDown) {
    return (
      <div
        role="alert"
        aria-live="assertive"
        className="animate-slide-down-banner"
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 9998,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: '8px',
          padding: '10px 16px',
          fontSize: '12px',
          fontWeight: 600,
          color: '#ffffff',
          background: 'linear-gradient(90deg, #f59e0b 0%, #d97706 100%)',
          boxShadow: '0 2px 8px rgba(245,158,11,0.4)',
          borderBottom: '1px solid #b45309',
        }}
      >
        <ServerCrash className="animate-pulse" style={{ width: 14, height: 14 }} />
        <span>Server unreachable — retrying automatically…</span>
      </div>
    )
  }

  if (showGreen) {
    return (
      <div
        role="status"
        aria-live="polite"
        className="animate-slide-down-banner"
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 9998,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: '8px',
          padding: '10px 16px',
          fontSize: '12px',
          fontWeight: 600,
          color: '#ffffff',
          background: 'linear-gradient(90deg, #059669 0%, #047857 100%)',
          boxShadow: '0 2px 8px rgba(5,150,105,0.4)',
          borderBottom: '1px solid #065f46',
          opacity: fadingOut ? 0 : 1,
          transition: 'opacity 0.3s ease-out',
        }}
      >
        <Server style={{ width: 14, height: 14 }} />
        <span>Server back online ✓</span>
      </div>
    )
  }

  return null
}
