import React from 'react'
import ReactDOM from 'react-dom/client'
import * as Sentry from '@sentry/react'
import App from './App'
import { ToastProvider } from './components/ui/Toast'
import { NetworkProvider } from './contexts/NetworkContext'
import { ErrorBoundary } from './components/ErrorBoundary'
import './index.css'

// Initialize Sentry BEFORE anything else
const SENTRY_DSN = import.meta.env.VITE_SENTRY_DSN
if (SENTRY_DSN) {
  Sentry.init({
    dsn: SENTRY_DSN,
    environment: import.meta.env.MODE || 'development',
    release: 'noant@2.0.0',
    tracesSampleRate: 0.2,
    replaysSessionSampleRate: 0.1,
    replaysOnErrorSampleRate: 1.0,
    integrations: [
      Sentry.browserTracingIntegration(),
      Sentry.replayIntegration({
        maskAllText: true,
        blockAllMedia: true,
      }),
    ],
    beforeSend(event) {
      // Scrub tokens from breadcrumbs
      if (event.breadcrumbs) {
        for (const crumb of event.breadcrumbs) {
          if (crumb.data) {
            delete crumb.data.token
            delete crumb.data.authorization
          }
        }
      }
      return event
    },
  })
}

// Initialize theme BEFORE React renders to prevent flash
function initTheme() {
  const saved = localStorage.getItem('noant_theme')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  
  // Always set ONE of them — never leave it to chance
  if (saved === 'dark' || (!saved && prefersDark)) {
    document.documentElement.classList.add('dark')
    document.documentElement.classList.remove('light')
  } else {
    document.documentElement.classList.add('light')
    document.documentElement.classList.remove('dark')
  }
}

initTheme()

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Sentry.ErrorBoundary fallback={<ErrorBoundary />}>
      <NetworkProvider>
        <ToastProvider>
          <App />
        </ToastProvider>
      </NetworkProvider>
    </Sentry.ErrorBoundary>
  </React.StrictMode>
)