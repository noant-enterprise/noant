import { createContext, useContext, useEffect, useRef, useState, useCallback, type ReactNode } from 'react'

interface NetworkContextValue {
  isOnline: boolean
  wasOffline: boolean
  isServerDown: boolean
  wasServerDown: boolean
  lastChecked: Date | null
}

const NetworkContext = createContext<NetworkContextValue>({
  isOnline: true,
  wasOffline: false,
  isServerDown: false,
  wasServerDown: false,
  lastChecked: null,
})

const RECONNECT_BANNER_DURATION = 2_500
const HEALTH_CHECK_INTERVAL = 15_000
const SERVER_DOWN_THRESHOLD = 3

function getHealthUrl(): string {
  const apiUrl = import.meta.env.VITE_API_URL as string | undefined
  if (apiUrl && apiUrl.startsWith('http')) {
    const base = apiUrl.replace(/\/api\/v1\/?$/, '').replace(/\/$/, '')
    return `${base}/health`
  }
  if (apiUrl && !apiUrl.startsWith('/')) {
    return `https://${apiUrl.replace(/\/api\/v1\/?$/, '').replace(/\/$/, '')}/health`
  }
  return '/health'
}

export function NetworkProvider({ children }: { children: ReactNode }) {
  const [isOnline, setIsOnline] = useState(
    typeof navigator !== 'undefined' ? navigator.onLine : true
  )
  const [wasOffline, setWasOffline] = useState(false)
  const [isServerDown, setIsServerDown] = useState(false)
  const [wasServerDown, setWasServerDown] = useState(false)
  const [lastChecked, setLastChecked] = useState<Date | null>(null)

  const prevOnline = useRef(isOnline)
  const prevServerDown = useRef(isServerDown)
  const consecutiveFailures = useRef(0)
  const failTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Browser online/offline
  useEffect(() => {
    if (typeof window === 'undefined') return
    const handleOnline = () => { setIsOnline(true); setLastChecked(new Date()) }
    const handleOffline = () => { setIsOnline(false); setLastChecked(new Date()) }
    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)
    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  // Flash "wasOffline" on offline→online
  useEffect(() => {
    if (!prevOnline.current && isOnline) {
      setWasOffline(true)
      const timer = setTimeout(() => setWasOffline(false), RECONNECT_BANNER_DURATION)
      prevOnline.current = isOnline
      return () => clearTimeout(timer)
    }
    prevOnline.current = isOnline
  }, [isOnline])

  // Flash "wasServerDown" on server down→up
  useEffect(() => {
    if (prevServerDown.current && !isServerDown) {
      setWasServerDown(true)
      const timer = setTimeout(() => setWasServerDown(false), RECONNECT_BANNER_DURATION)
      prevServerDown.current = isServerDown
      return () => clearTimeout(timer)
    }
    prevServerDown.current = isServerDown
  }, [isServerDown])

  const checkHealth = useCallback(async () => {
    try {
      const res = await fetch(getHealthUrl(), {
        method: 'GET',
        signal: AbortSignal.timeout(5000),
      })
      if (res.ok) {
        consecutiveFailures.current = 0
        setIsServerDown(false)
      } else {
        consecutiveFailures.current++
        if (consecutiveFailures.current >= SERVER_DOWN_THRESHOLD) {
          setIsServerDown(true)
        }
      }
    } catch {
      consecutiveFailures.current++
      if (consecutiveFailures.current >= SERVER_DOWN_THRESHOLD) {
        setIsServerDown(true)
      }
    }
    setLastChecked(new Date())
  }, [])

  // Periodic health check
  useEffect(() => {
    checkHealth()
    const interval = setInterval(checkHealth, HEALTH_CHECK_INTERVAL)
    return () => {
      clearInterval(interval)
      if (failTimer.current) clearTimeout(failTimer.current)
    }
  }, [checkHealth])

  return (
    <NetworkContext.Provider value={{ isOnline, wasOffline, isServerDown, wasServerDown, lastChecked }}>
      {children}
    </NetworkContext.Provider>
  )
}

export function useNetwork() {
  return useContext(NetworkContext)
}
