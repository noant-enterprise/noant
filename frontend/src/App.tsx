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
import LandingPage from '@/app/landing/page'
import { useEffect } from 'react'
import { OfflineBanner } from '@/components/OfflineBanner'
import { NetworkProvider } from '@/contexts/NetworkContext'
import { refreshToken } from '@/lib/auth'

function AppShell() {
  const location = useLocation()

  useEffect(() => {
    const authPaths = ['/login', '/signup', '/forgot-password', '/reset-password']
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
        element: <LandingPage />,
      },
      // Auth routes - guest protection
      {
        path: '/login',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <LoginPage /> }],
      },
      {
        path: '/signup',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <SignupPage /> }],
      },
      {
        path: '/forgot-password',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <ForgotPasswordPage /> }],
      },
      {
        path: '/reset-password',
        element: (
          <GuestRoute>
            <AuthLayout />
          </GuestRoute>
        ),
        children: [{ index: true, element: <ResetPasswordPage /> }],
      },
      // Dashboard routes - protected
      {
        path: '/',
        element: (
          <ProtectedRoute>
            <WidgetConfigProvider>
              <SidebarAlertsProvider>
                <DashboardLayout />
              </SidebarAlertsProvider>
            </WidgetConfigProvider>
          </ProtectedRoute>
        ),
        children: [
          { path: 'dashboard', element: <OverviewPage /> },
          { path: 'chats', element: <ChatsPage /> },
          { path: 'teach', element: <TeachPage /> },
          { path: 'insights', element: <InsightsPage /> },
          { path: 'channels', element: <ChannelsPage /> },

          { path: 'settings', element: <SettingsPage /> },
          { path: 'notifications', element: <NotificationsPage /> },
          { path: 'billing', element: <BillingPage /> },
          { path: 'team', element: <TeamPage /> },
          { path: 'widget', element: <WidgetPage /> },
          { path: 'leads', element: <LeadsPage /> },
          { path: 'inventory', element: <InventoryPage /> },
        ],
      },
      { path: '*', element: <Navigate to='/' replace /> },
    ],
  },
])

import { WidgetConfigProvider } from '@/contexts/WidgetConfigContext';
import { SidebarAlertsProvider } from '@/contexts/SidebarAlertsContext';

export default function App() {
  return (
    <NetworkProvider>
      <RouterProvider router={router} />
    </NetworkProvider>
  );
}
