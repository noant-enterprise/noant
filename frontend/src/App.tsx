import { createBrowserRouter, RouterProvider, Navigate, Outlet, useLocation } from 'react-router-dom'
import { Suspense, lazy, useEffect } from 'react'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { DashboardLayout } from '@/components/layout/DashboardLayout'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'
import { GuestRoute } from '@/components/auth/GuestRoute'
import { CommandPalette } from '@/components/ui/CommandPalette'
import { OfflineBanner } from '@/components/OfflineBanner'
import { ServerDownBanner } from '@/components/ServerDownBanner'
import { refreshToken } from '@/lib/auth'
import { PwaInstallPrompt } from '@/components/ui/PwaInstallPrompt'
import { WidgetConfigProvider } from '@/contexts/WidgetConfigContext'
import { SidebarAlertsProvider } from '@/contexts/SidebarAlertsContext'
import { ModalProvider } from '@/contexts/ModalContext'
import { RouteErrorBoundary } from '@/components/RouteErrorBoundary'
import { Spinner } from '@/components/ui/Spinner'
import { useToast } from '@/components/ui/Toast'

const LoginPage = lazy(() => import('@/app/(auth)/login/page'))
const SignupPage = lazy(() => import('@/app/(auth)/signup/page'))
const ForgotPasswordPage = lazy(() => import('@/app/(auth)/forgot-password/page'))
const ResetPasswordPage = lazy(() => import('@/app/(auth)/reset-password/page'))
const VerifyEmailPage = lazy(() => import('@/app/(auth)/verify-email/page'))
const OverviewPage = lazy(() => import('@/app/(dashboard)/page'))
const ChatsPage = lazy(() => import('@/app/(dashboard)/chats/page'))
const TeachPage = lazy(() => import('@/app/(dashboard)/teach/page'))
const InsightsPage = lazy(() => import('@/app/(dashboard)/insights/page'))
const ChannelsPage = lazy(() => import('@/app/(dashboard)/channels/page'))
const SettingsPage = lazy(() => import('@/app/(dashboard)/settings/page'))
const NotificationsPage = lazy(() => import('@/app/(dashboard)/notifications/page'))
const BillingPage = lazy(() => import('@/app/(dashboard)/billing/page'))
const TeamPage = lazy(() => import('@/app/(dashboard)/team/page'))
const WidgetPage = lazy(() => import('@/app/(dashboard)/widget/page'))
const LeadsPage = lazy(() => import('@/app/(dashboard)/leads/page'))
const InventoryPage = lazy(() => import('@/app/(dashboard)/inventory/page'))
const OnboardingPage = lazy(() => import('@/app/(dashboard)/onboarding/page'))
const UnknownQuestionsPage = lazy(() => import('@/app/(dashboard)/teach/unknown/page'))
const LandingPage = lazy(() => import('@/app/landing/page'))

function PageLoader() {
  return (
    <div className="flex items-center justify-center h-screen w-full" style={{ background: 'var(--bg-base)' }}>
      <Spinner size="lg" className="text-tertiary" />
    </div>
  )
}

function AppShell() {
  const location = useLocation()
  const { toast } = useToast()

  useEffect(() => {
    const authPaths = ['/login', '/signup', '/forgot-password', '/reset-password', '/verify-email']
    if (authPaths.some((path) => location.pathname.startsWith(path))) {
      return
    }

    const refreshSession = async () => {
      try {
        await refreshToken()
      } catch (err) {
        console.error('Failed to refresh session in background:', err)
        toast('Session expired. Please log in again.', 'error')
      }
    }

    refreshSession()

    const interval = setInterval(refreshSession, 20 * 60 * 1000)
    return () => clearInterval(interval);
  }, [location.pathname, toast]);

  return (
    <>
      <OfflineBanner />
      <ServerDownBanner />
      <CommandPalette />
      <PwaInstallPrompt />
      <Outlet />
    </>
  )
}

const router = createBrowserRouter([
  {
    element: <AppShell />,
    children: [
      {
        path: '/',
        element: (
          <RouteErrorBoundary pageName="Home">
            <Suspense fallback={<PageLoader />}>
              <LandingPage />
            </Suspense>
          </RouteErrorBoundary>
        ),
      },
      {
        path: '/login',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <RouteErrorBoundary pageName="Login"><Suspense fallback={<PageLoader />}><LoginPage /></Suspense></RouteErrorBoundary> }],
      },
      {
        path: '/signup',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <RouteErrorBoundary pageName="Signup"><Suspense fallback={<PageLoader />}><SignupPage /></Suspense></RouteErrorBoundary> }],
      },
      {
        path: '/forgot-password',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <RouteErrorBoundary pageName="Forgot Password"><Suspense fallback={<PageLoader />}><ForgotPasswordPage /></Suspense></RouteErrorBoundary> }],
      },
      {
        path: '/reset-password',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <RouteErrorBoundary pageName="Reset Password"><Suspense fallback={<PageLoader />}><ResetPasswordPage /></Suspense></RouteErrorBoundary> }],
      },
      {
        path: '/verify-email',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <RouteErrorBoundary pageName="Verify Email"><Suspense fallback={<PageLoader />}><VerifyEmailPage /></Suspense></RouteErrorBoundary> }],
      },
      {
        path: '/',
        element: (
          <ProtectedRoute>
            <WidgetConfigProvider>
              <SidebarAlertsProvider>
                <RouteErrorBoundary pageName="Dashboard">
                  <DashboardLayout />
                </RouteErrorBoundary>
              </SidebarAlertsProvider>
            </WidgetConfigProvider>
          </ProtectedRoute>
        ),
        children: [
          { path: 'dashboard', element: <RouteErrorBoundary pageName="Overview"><Suspense fallback={<PageLoader />}><OverviewPage /></Suspense></RouteErrorBoundary> },
          { path: 'chats', element: <RouteErrorBoundary pageName="Chats"><Suspense fallback={<PageLoader />}><ChatsPage /></Suspense></RouteErrorBoundary> },
          { path: 'teach', children: [
            { index: true, element: <RouteErrorBoundary pageName="Teach"><Suspense fallback={<PageLoader />}><TeachPage /></Suspense></RouteErrorBoundary> },
            { path: 'unknown', element: <RouteErrorBoundary pageName="Unknown Questions"><Suspense fallback={<PageLoader />}><UnknownQuestionsPage /></Suspense></RouteErrorBoundary> },
          ]},
          { path: 'insights', element: <RouteErrorBoundary pageName="Insights"><Suspense fallback={<PageLoader />}><InsightsPage /></Suspense></RouteErrorBoundary> },
          { path: 'channels', element: <RouteErrorBoundary pageName="Channels"><Suspense fallback={<PageLoader />}><ChannelsPage /></Suspense></RouteErrorBoundary> },
          { path: 'settings', element: <RouteErrorBoundary pageName="Settings"><Suspense fallback={<PageLoader />}><SettingsPage /></Suspense></RouteErrorBoundary> },
          { path: 'notifications', element: <RouteErrorBoundary pageName="Notifications"><Suspense fallback={<PageLoader />}><NotificationsPage /></Suspense></RouteErrorBoundary> },
          { path: 'billing', element: <RouteErrorBoundary pageName="Billing"><Suspense fallback={<PageLoader />}><BillingPage /></Suspense></RouteErrorBoundary> },
          { path: 'team', element: <RouteErrorBoundary pageName="Team"><Suspense fallback={<PageLoader />}><TeamPage /></Suspense></RouteErrorBoundary> },
          { path: 'widget', element: <RouteErrorBoundary pageName="Widget"><Suspense fallback={<PageLoader />}><WidgetPage /></Suspense></RouteErrorBoundary> },
          { path: 'leads', element: <RouteErrorBoundary pageName="Leads"><Suspense fallback={<PageLoader />}><LeadsPage /></Suspense></RouteErrorBoundary> },
          { path: 'inventory', element: <RouteErrorBoundary pageName="Inventory"><Suspense fallback={<PageLoader />}><InventoryPage /></Suspense></RouteErrorBoundary> },
          { path: 'onboarding', element: <RouteErrorBoundary pageName="Onboarding"><Suspense fallback={<PageLoader />}><OnboardingPage /></Suspense></RouteErrorBoundary> },
        ],
      },
      { path: '*', element: <Navigate to='/' replace /> },
    ],
  },
])

export default function App() {
  return (
    <ModalProvider>
      <RouterProvider router={router} />
    </ModalProvider>
  )
}
