import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import { ToastProvider } from './components/ui/Toast'
import { ErrorBoundary } from './components/ErrorBoundary'
import './index.css'

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
    <ErrorBoundary>
      <ToastProvider>
        <App />
      </ToastProvider>
    </ErrorBoundary>
  </React.StrictMode>
)