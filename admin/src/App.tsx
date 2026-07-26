import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Shell } from '@/components/layout/Shell'
import { useAuth } from '@/lib/hooks/useAuth'
import LoginPage from '@/app/login/page'
import DashboardPage from '@/app/page'
import CustomersPage from '@/app/customers/page'
import CustomerDetailPage from '@/app/customers/[id]/page'
import AnalyticsPage from '@/app/analytics/page'
import RevenuePage from '@/app/revenue/page'
import AIHealthPage from '@/app/ai/page'
import SystemPage from '@/app/system/page'
import SettingsPage from '@/app/settings/page'
import AuditLogsPage from '@/app/audit-logs/page'
import KnowledgeBasePage from '@/app/knowledge/page'
import PipelinePage from '@/app/pipeline/page'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, loading } = useAuth()
  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-bg-base">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-brand-sky border-t-transparent" />
      </div>
    )
  }
  if (!isAuthenticated) return <Navigate to="/login" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          element={
            <ProtectedRoute>
              <Shell />
            </ProtectedRoute>
          }
        >
          <Route path="/" element={<DashboardPage />} />
          <Route path="/customers" element={<CustomersPage />} />
          <Route path="/customers/:id" element={<CustomerDetailPage />} />
          <Route path="/analytics" element={<AnalyticsPage />} />
          <Route path="/revenue" element={<RevenuePage />} />
          <Route path="/ai" element={<AIHealthPage />} />
          <Route path="/system" element={<SystemPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/audit-logs" element={<AuditLogsPage />} />
          <Route path="/knowledge" element={<KnowledgeBasePage />} />
          <Route path="/pipeline" element={<PipelinePage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
