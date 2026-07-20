import { Component, type ReactNode } from 'react'
import { AlertTriangle, RefreshCw, ArrowLeft } from 'lucide-react'
import { Button } from './ui/Button'

interface Props {
  children: ReactNode
  pageName?: string
}

interface State {
  hasError: boolean
  error?: Error
}

export class RouteErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error(`RouteErrorBoundary (${this.props.pageName || 'page'}):`, error, info.componentStack)
  }

  handleTryAgain = () => {
    this.setState({ hasError: false, error: undefined })
  }

  handleReload = () => {
    window.location.reload()
  }

  handleGoBack = () => {
    window.history.back()
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-[60vh] flex items-center justify-center p-6">
          <div className="max-w-sm w-full text-center">
            <div className="w-12 h-12 bg-red-50 dark:bg-red-900/20 rounded-xl flex items-center justify-center mx-auto mb-4">
              <AlertTriangle className="w-6 h-6 text-red-500" />
            </div>
            <h2 className="text-lg font-semibold text-primary mb-1">
              {this.props.pageName || 'Page'} failed to load
            </h2>
            <p className="text-sm text-secondary mb-4">
              An unexpected error occurred while rendering this page.
            </p>
            <div className="bg-inset rounded-lg p-3 mb-4 text-left">
              <code className="text-xs text-red-500 font-mono break-all">
                {this.state.error?.message || 'Unknown error'}
              </code>
            </div>
            <div className="flex gap-2 justify-center">
              <Button onClick={this.handleTryAgain} className="gap-2 text-sm">
                <RefreshCw className="w-3 h-3" />
                Try again
              </Button>
              <Button variant="ghost" onClick={this.handleGoBack} className="gap-2 text-sm">
                <ArrowLeft className="w-3 h-3" />
                Go back
              </Button>
            </div>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
