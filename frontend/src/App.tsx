import { createBrowserRouter, RouterProvider, Navigate, Outlet, useLocation } from 'react-router-dom'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { DashboardLayout } from '@/components/layout/DashboardLayout'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'
import { GuestRoute } from '@/components/auth/GuestRoute'
import { CommandPalette } from '@/components/ui/CommandPalette'
import LoginPage from '@/app/(auth)/login/page'
import SignupPage from '@/app/(auth)/signup/page'
import ForgotPasswordPage from '@/app/(auth)/forgot-password/page'
import ResetPasswordPage from '@/app/(auth)/reset-password/page'
import VerifyEmailPage from '@/app/(auth)/verify-email/page'
import OverviewPage from '@/app/(dashboard)/page'
import ChatsPage from '@/app/(dashboard)/chats/page'
import TeachPage from '@/app/(dashboard)/teach/page'
import InsightsPage from '@/app/(dashboard)/insights/page'
import ChannelsPage from '@/app/(dashboard)/channels/page'

import SettingsPage from '@/app/(dashboard)/settings/page'
import NotificationsPage from '@/app/(dashboard)/notifications/page'
import BillingPage from '@/app/(dashboard)/billing/page'
import TeamPage from '@/app/(dashboard)/team/page'
import WidgetPage from '@/app/(dashboard)/widget/page'
import LeadsPage from '@/app/(dashboard)/leads/page'
import InventoryPage from '@/app/(dashboard)/inventory/page'
import OnboardingPage from '@/app/(dashboard)/onboarding/page'
import UnknownQuestionsPage from '@/app/(dashboard)/teach/unknown/page'
import LandingPage from '@/app/landing/page'
import { useEffect } from 'react'
import { OfflineBanner } from '@/components/OfflineBanner'
import { refreshToken } from '@/lib/auth'
import { PwaInstallPrompt } from '@/components/ui/PwaInstallPrompt'

function AppShell() {
  const location = useLocation()

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
      }
    }

    refreshSession()

    const interval = setInterval(refreshSession, 20 * 60 * 1000)
    return () => clearInterval(interval);
  }, [location.pathname]);

  return (
    <>
      <OfflineBanner />
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
      // Public landing page
      {
        path: '/',
        element: (
          <RouteErrorBoundary pageName="Home">
            <LandingPage />
          </RouteErrorBoundary>
        ),
      },
      // Auth routes - guest protection
      {
        path: '/login',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <RouteErrorBoundary pageName="Login"><LoginPage /></RouteErrorBoundary> }],
      },
      {
        path: '/signup',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <RouteErrorBoundary pageName="Signup"><SignupPage /></RouteErrorBoundary> }],
      },
      {
        path: '/forgot-password',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <RouteErrorBoundary pageName="Forgot Password"><ForgotPasswordPage /></RouteErrorBoundary> }],
      },
      {
        path: '/reset-password',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <RouteErrorBoundary pageName="Reset Password"><ResetPasswordPage /></RouteErrorBoundary> }],
      },
      {
        path: '/verify-email',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <RouteErrorBoundary pageName="Verify Email"><VerifyEmailPage /></RouteErrorBoundary> }],
      },
      // Dashboard routes - protected
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
          { path: 'dashboard', element: <RouteErrorBoundary pageName="Overview"><OverviewPage /></RouteErrorBoundary> },
          { path: 'chats', element: <RouteErrorBoundary pageName="Chats"><ChatsPage /></RouteErrorBoundary> },
          { path: 'teach', children: [
            { index: true, element: <RouteErrorBoundary pageName="Teach"><TeachPage /></RouteErrorBoundary> },
            { path: 'unknown', element: <RouteErrorBoundary pageName="Unknown Questions"><UnknownQuestionsPage /></RouteErrorBoundary> },
          ]},
          { path: 'insights', element: <RouteErrorBoundary pageName="Insights"><InsightsPage /></RouteErrorBoundary> },
          { path: 'channels', element: <RouteErrorBoundary pageName="Channels"><ChannelsPage /></RouteErrorBoundary> },

          { path: 'settings', element: <RouteErrorBoundary pageName="Settings"><SettingsPage /></RouteErrorBoundary> },
          { path: 'notifications', element: <RouteErrorBoundary pageName="Notifications"><NotificationsPage /></RouteErrorBoundary> },
          { path: 'billing', element: <RouteErrorBoundary pageName="Billing"><BillingPage /></RouteErrorBoundary> },
          { path: 'team', element: <RouteErrorBoundary pageName="Team"><TeamPage /></RouteErrorBoundary> },
          { path: 'widget', element: <RouteErrorBoundary pageName="Widget"><WidgetPage /></RouteErrorBoundary> },
          { path: 'leads', element: <RouteErrorBoundary pageName="Leads"><LeadsPage /></RouteErrorBoundary> },
          { path: 'inventory', element: <RouteErrorBoundary pageName="Inventory"><InventoryPage /></RouteErrorBoundary> },
          { path: 'onboarding', element: <RouteErrorBoundary pageName="Onboarding"><OnboardingPage /></RouteErrorBoundary> },
        ],
      },
      { path: '*', element: <Navigate to='/' replace /> },
    ],
  },
])

import { WidgetConfigProvider } from '@/contexts/WidgetConfigContext';
import { SidebarAlertsProvider } from '@/contexts/SidebarAlertsContext';
import { ModalProvider } from '@/contexts/ModalContext';
import { RouteErrorBoundary } from '@/components/RouteErrorBoundary';

export default function App() {
  return (
    <ModalProvider>
      <RouterProvider router={router} />
    </ModalProvider>
  );
}
