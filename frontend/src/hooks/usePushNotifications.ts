import { useState, useEffect, useCallback } from 'react'
import { api } from '@/lib/api'

const VAPID_PUBLIC_KEY = import.meta.env.VITE_VAPID_PUBLIC_KEY || ''

function urlBase64ToUint8Array(base64: string): Uint8Array {
  const padding = '='.repeat((4 - (base64.length % 4)) % 4)
  const b64 = (base64 + padding).replace(/-/g, '+').replace(/_/g, '/')
  const raw = atob(b64)
  const arr = new Uint8Array(raw.length)
  for (let i = 0; i < raw.length; i++) {
    arr[i] = raw.charCodeAt(i)
  }
  return arr
}

export function usePushNotifications() {
  const [permission, setPermission] = useState<NotificationPermission>('default')
  const [subscribed, setSubscribed] = useState(false)
  const [loading, setLoading] = useState(false)
  const [supported, setSupported] = useState(false)

  useEffect(() => {
    setSupported('serviceWorker' in navigator && 'PushManager' in window)
    if ('Notification' in window) {
      setPermission(Notification.permission)
    }
    checkSubscription()
  }, [])

  const checkSubscription = useCallback(async () => {
    try {
      const reg = await navigator.serviceWorker.ready
      const sub = await reg.pushManager.getSubscription()
      setSubscribed(!!sub)
    } catch {
      setSubscribed(false)
    }
  }, [])

  const subscribe = useCallback(async () => {
    if (!supported || !VAPID_PUBLIC_KEY) return
    setLoading(true)
    try {
      const permission = await Notification.requestPermission()
      setPermission(permission)
      if (permission !== 'granted') return

      const reg = await navigator.serviceWorker.ready
      let sub = await reg.pushManager.getSubscription()
      if (!sub) {
        const key = urlBase64ToUint8Array(VAPID_PUBLIC_KEY) as unknown as BufferSource
        sub = await reg.pushManager.subscribe({
          userVisibleOnly: true,
          applicationServerKey: key,
        })
      }

      const json = sub.toJSON()
      await api.post('/push/subscribe', {
        endpoint: sub.endpoint,
        auth: json.keys?.auth ?? '',
        p256dh: json.keys?.p256dh ?? '',
      })

      setSubscribed(true)
    } catch (e) {
      console.error('Push subscription failed:', e)
    } finally {
      setLoading(false)
    }
  }, [supported])

  const unsubscribe = useCallback(async () => {
    setLoading(true)
    try {
      const reg = await navigator.serviceWorker.ready
      const sub = await reg.pushManager.getSubscription()
      const endpoint = sub?.endpoint ?? ''
      if (sub) {
        await sub.unsubscribe()
      }
      await api.post('/push/unsubscribe', { endpoint })
      setSubscribed(false)
    } catch (e) {
      console.error('Push unsubscription failed:', e)
    } finally {
      setLoading(false)
    }
  }, [])

  return { supported, permission, subscribed, loading, subscribe, unsubscribe }
}
