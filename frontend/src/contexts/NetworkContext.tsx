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

const RECONNECT_BANNER_DURATION = 2_500  // 2.5 s

export function NetworkProvider({ children }: { children: ReactNode }) {
  const [isOnline, setIsOnline] = useState(
    typeof navigator !== 'undefined' ? navigator.onLine : true
  )
  const [wasOffline, setWasOffline] = useState(false)
  const [lastChecked, setLastChecked] = useState<Date | null>(null)

  // Track whether we were previously offline so we can flash "Back online" banner
  const prevOnline = useRef(isOnline)

  useEffect(() => {
    if (typeof window === 'undefined') return

    const handleOnline = () => {
      setIsOnline(true)
      setLastChecked(new Date())
    }

    const handleOffline = () => {
      setIsOnline(false)
      setLastChecked(new Date())
    }

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  // Flash "wasOffline" when transitioning offline → online
  useEffect(() => {
    if (!prevOnline.current && isOnline) {
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
