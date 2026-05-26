import { useEffect, useRef, useState } from 'react'
import { WifiOff, Wifi } from 'lucide-react'
import { useNetwork } from '@/contexts/NetworkContext'

/**
 * OfflineBanner — sticky top banner that:
 *  - Shows red "You are offline" when network is down (slides down instantly)
 *  - Flashes green "Back online" for 2.5 s after reconnecting, then fades out
 *  - z-index 9999 (below modals at 10000, above toasts at 9000)
 */
export function OfflineBanner() {
  const { isOnline, wasOffline } = useNetwork()

  // Local state to drive CSS fade-out before unmounting the green banner
  const [showGreen, setShowGreen] = useState(false)
  const [fadingOut, setFadingOut] = useState(false)
  const fadeTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (wasOffline && isOnline) {
      // Start showing green banner
      setShowGreen(true)
      setFadingOut(false)

      // Begin fade-out 300ms before wasOffline resets (2500ms - 300ms = 2200ms)
      fadeTimer.current = setTimeout(() => {
        setFadingOut(true)
        // Unmount after fade finishes (300ms)
        setTimeout(() => setShowGreen(false), 300)
      }, 2200)
    } else if (!wasOffline) {
      setShowGreen(false)
      setFadingOut(false)
    }

    return () => {
      if (fadeTimer.current) clearTimeout(fadeTimer.current)
    }
  }, [wasOffline, isOnline])

  // Offline state — red banner
  if (!isOnline) {
    return (
      <div
        role="alert"
        aria-live="assertive"
        className="animate-slide-down-banner"
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 9999,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: '8px',
          padding: '10px 16px',
          fontSize: '12px',
          fontWeight: 600,
          color: '#ffffff',
          background: 'linear-gradient(90deg, #dc2626 0%, #b91c1c 100%)',
          boxShadow: '0 2px 8px rgba(220,38,38,0.4)',
          borderBottom: '1px solid #991b1b',
        }}
      >
        <WifiOff className="animate-pulse" style={{ width: 14, height: 14 }} />
        <span>You're offline — trying to reconnect…</span>
      </div>
    )
  }

  // Online-after-offline state — green banner with fade out
  if (showGreen) {
    return (
      <div
        role="status"
        aria-live="polite"
        className="animate-slide-down-banner"
        style={{
          position: 'sticky',
          top: 0,
          zIndex: 9999,
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
        <Wifi style={{ width: 14, height: 14 }} />
        <span>Back online ✓</span>
      </div>
    )
  }

  return null
}
