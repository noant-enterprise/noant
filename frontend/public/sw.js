const CACHE = 'noant-v1'

const ASSETS = self.__WB_MANIFEST || []

self.addEventListener('install', (event) => {
  self.skipWaiting()
  event.waitUntil(
    caches.open(CACHE).then((cache) => cache.addAll(ASSETS.map((a) => a.url)))
  )
})

self.addEventListener('activate', (event) => {
  event.waitUntil(clients.claim())
})

self.addEventListener('fetch', (event) => {
  if (event.request.method !== 'GET') return
  if (event.request.url.includes('/api/')) {
    event.respondWith(networkFirst(event.request))
    return
  }
  event.respondWith(cacheFirst(event.request))
})

async function networkFirst(request) {
  try {
    const resp = await fetch(request)
    if (resp.ok) {
      const cache = await caches.open(CACHE)
      cache.put(request, resp.clone())
    }
    return resp
  } catch {
    const cached = await caches.match(request)
    return cached || new Response('Offline', { status: 503 })
  }
}

async function cacheFirst(request) {
  const cached = await caches.match(request)
  if (cached) return cached
  try {
    const resp = await fetch(request)
    if (resp.ok) {
      const cache = await caches.open(CACHE)
      cache.put(request, resp.clone())
    }
    return resp
  } catch {
    return new Response('Offline', { status: 503 })
  }
}

self.addEventListener('push', (event) => {
  if (!event.data) return
  try {
    const data = event.data.json()
    const title = data.title || 'Noant'
    const options = {
      body: data.body || '',
      icon: '/favicon.jpg',
      badge: '/favicon.jpg',
      data: { url: data.url || '/' },
      vibrate: [200, 100, 200],
    }
    event.waitUntil(self.registration.showNotification(title, options))
  } catch {
    const title = 'Noant'
    const options = { body: event.data.text(), icon: '/favicon.jpg' }
    event.waitUntil(self.registration.showNotification(title, options))
  }
})

self.addEventListener('notificationclick', (event) => {
  event.notification.close()
  const url = event.notification.data?.url || '/'
  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
      for (const client of clientList) {
        if (client.url.startsWith(self.location.origin) && 'focus' in client) {
          client.postMessage({ type: 'navigate', url })
          return client.focus()
        }
      }
      if (clients.openWindow) return clients.openWindow(url)
    })
  )
})
