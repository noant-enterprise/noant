import { Component, type ReactNode } from 'react'
import { AlertTriangle, RefreshCw } from 'lucide-react'
import { Button } from './ui/Button'

interface Props {
  children?: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error('ErrorBoundary caught:', error, info.componentStack)
  }

  handleTryAgain = () => {
    this.setState({ hasError: false, error: undefined })
  }

  handleReload = () => {
    window.location.reload()
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      return (
        <div className="min-h-screen flex items-center justify-center bg-base p-6">
          <div className="max-w-md w-full text-center">
            <div className="w-16 h-16 bg-red-50 dark:bg-red-900/20 rounded-2xl flex items-center justify-center mx-auto mb-6">
              <AlertTriangle className="w-8 h-8 text-red-500" />
            </div>
            <h1 className="text-2xl font-bold text-primary mb-2">Something went wrong</h1>
            <p className="text-secondary mb-6">
              An unexpected error occurred. You can try again or reload the page.
            </p>
            <div className="bg-inset rounded-lg p-4 mb-6 text-left">
              <code className="text-xs text-red-500 font-mono break-all">
                {this.state.error?.message || 'Unknown error'}
              </code>
            </div>
            <div className="flex gap-3 justify-center">
              <Button onClick={this.handleTryAgain} className="gap-2">
                <RefreshCw className="w-4 h-4" />
                Try again
              </Button>
              <Button variant="ghost" onClick={this.handleReload} className="gap-2">
                Reload page
              </Button>
            </div>
          </div>
        </div>
      )
    }

    return this.props.children ?? null
  }
}