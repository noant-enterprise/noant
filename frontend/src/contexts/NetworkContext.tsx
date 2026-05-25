import { createContext, useContext, useEffect, useRef, useState, ReactNode } from 'react'

interface NetworkContextValue {
  isOnline: boolean
  wasOffline: boolean        // true briefly after reconnecting (for "Back online" banner)
  lastChecked: Date | null
}

const NetworkContext = createContext<NetworkContextValue>({
  isOnline: true,
  wasOffline: false,
  lastChecked: null,
})

const HEALTH_PING_URL = '/api/health'
const PING_INTERVAL_MS = 30_000   // 30 s
const RECONNECT_BANNER_DURATION = 2_500  // 2.5 s

export function NetworkProvider({ children }: { children: ReactNode }) {
  const [isOnline, setIsOnline] = useState(
    typeof navigator !== 'undefined' ? navigator.onLine : true
  )
  const [wasOffline, setWasOffline] = useState(false)
  const [lastChecked, setLastChecked] = useState<Date | null>(null)

  // Track whether we were previously offline so we can flash the "Back online" banner
  const prevOnline = useRef(isOnline)

  const ping = async () => {
    try {
      const res = await fetch(HEALTH_PING_URL, {
        method: 'GET',
        cache: 'no-store',
        signal: AbortSignal.timeout(5_000),
      })
      const alive = res.ok
      setIsOnline(alive)
      setLastChecked(new Date())
      return alive
    } catch {
      setIsOnline(false)
      setLastChecked(new Date())
      return false
    }
  }

  useEffect(() => {
    if (typeof window === 'undefined') return

    const handleOnline = () => {
      setIsOnline(true)
      // ping to confirm we have real connectivity
      ping()
    }

    const handleOffline = () => {
      setIsOnline(false)
    }

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    // Periodic health ping while mounted
    const interval = setInterval(ping, PING_INTERVAL_MS)

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
      clearInterval(interval)
    }
  }, [])

  // Flash "wasOffline" when transitioning offline → online
  useEffect(() => {
    if (!prevOnline.current && isOnline) {
      // Just came back online
      setWasOffline(true)
      const timer = setTimeout(() => setWasOffline(false), RECONNECT_BANNER_DURATION)
      prevOnline.current = isOnline
      return () => clearTimeout(timer)
    }
    prevOnline.current = isOnline
  }, [isOnline])

  return (
    <NetworkContext.Provider value={{ isOnline, wasOffline, lastChecked }}>
      {children}
    </NetworkContext.Provider>
  )
}

export function useNetwork() {
  return useContext(NetworkContext)
}
