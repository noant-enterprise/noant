import { WifiOff, Wifi } from 'lucide-react'
import { useNetwork } from '@/contexts/NetworkContext'

/**
 * OfflineBanner — sticky top banner that:
 *  - Shows red "You are offline" when network is down
 *  - Flashes green "Back online" for 2.5 s after reconnecting
 *  - Slides in/out smoothly
 *  - z-index 9999 (below modals at 10000, above toasts at 9000)
 */
export function OfflineBanner() {
  const { isOnline, wasOffline } = useNetwork()

  // Determine which banner variant to show
  const showOffline = !isOnline
  const showOnline = isOnline && wasOffline

  const visible = showOffline || showOnline

  if (!visible) return null

  if (showOnline) {
    return (
      <div
        role="status"
        aria-live="polite"
        className="sticky top-0 z-[9999] flex items-center justify-center gap-2 py-2.5 px-4 text-xs font-semibold text-white bg-emerald-500 shadow-md border-b border-emerald-600 animate-slide-down"
      >
        <Wifi className="w-4 h-4" />
        <span>Back online</span>
      </div>
    )
  }

  return (
    <div
      role="alert"
      aria-live="assertive"
      className="sticky top-0 z-[9999] flex items-center justify-center gap-2 py-2.5 px-4 text-xs font-semibold text-white bg-red-600 shadow-md border-b border-red-700 animate-slide-down"
    >
      <WifiOff className="w-4 h-4 animate-pulse" />
      <span>You are currently offline. Trying to reconnect…</span>
    </div>
  )
}
