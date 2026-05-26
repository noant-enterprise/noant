import { createBrowserRouter, RouterProvider, Navigate, Outlet } from 'react-router-dom'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { DashboardLayout } from '@/components/layout/DashboardLayout'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'
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
import SetupPage from '@/app/(dashboard)/setup/page'
import SettingsPage from '@/app/(dashboard)/settings/page'
import NotificationsPage from '@/app/(dashboard)/notifications/page'
import BillingPage from '@/app/(dashboard)/billing/page'
import TeamPage from '@/app/(dashboard)/team/page'
import WidgetPage from '@/app/(dashboard)/widget/page'

import { useEffect } from 'react'
import { OfflineBanner } from '@/components/OfflineBanner'
import { NetworkProvider } from '@/contexts/NetworkContext'

function parseJwt(token: string) {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      window.atob(base64)
        .split('')
        .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );
    return JSON.parse(jsonPayload);
  } catch (e) {
    return null;
  }
}

function AppShell() {
  useEffect(() => {
    const checkAndRefreshToken = async () => {
      const token = localStorage.getItem('noant_token');
      if (!token) return;

      const claims = parseJwt(token);
      if (!claims || !claims.exp) return;

      const expiryTime = claims.exp * 1000;
      const now = Date.now();
      const timeLeft = expiryTime - now;

      // If token expires in less than 5 minutes, trigger a refresh
      if (timeLeft > 0 && timeLeft < 5 * 60 * 1000) {
        try {
          const { refreshToken } = await import('@/lib/auth');
          await refreshToken();
        } catch (err) {
          console.error('Failed to refresh token in background:', err);
        }
      }
    };

    // Run check on mount
    checkAndRefreshToken();

    // Check every minute
    const interval = setInterval(checkAndRefreshToken, 60 * 1000);
    return () => clearInterval(interval);
  }, []);

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
      // Auth routes - no protection needed
      {
        path: '/login',
        element: <AuthLayout />,
        children: [{ index: true, element: <LoginPage /> }],
      },
      {
        path: '/signup',
        element: <AuthLayout />,
        children: [{ index: true, element: <SignupPage /> }],
      },
      {
        path: '/forgot-password',
        element: <AuthLayout />,
        children: [{ index: true, element: <ForgotPasswordPage /> }],
      },
      {
        path: '/reset-password',
        element: <AuthLayout />,
        children: [{ index: true, element: <ResetPasswordPage /> }],
      },
      // Dashboard routes - protected
      {
        path: '/',
        element: (
          <ProtectedRoute>
            <WidgetConfigProvider>
              <DashboardLayout />
            </WidgetConfigProvider>
          </ProtectedRoute>
        ),
        children: [
          { index: true, element: <OverviewPage /> },
          { path: 'chats', element: <ChatsPage /> },
          { path: 'teach', element: <TeachPage /> },
          { path: 'insights', element: <InsightsPage /> },
          { path: 'channels', element: <ChannelsPage /> },
          { path: 'setup', element: <SetupPage /> },
          { path: 'settings', element: <SettingsPage /> },
          { path: 'notifications', element: <NotificationsPage /> },
          { path: 'billing', element: <BillingPage /> },
          { path: 'team', element: <TeamPage /> },
          { path: 'widget', element: <WidgetPage /> },
        ],
      },
      { path: '*', element: <Navigate to='/' replace /> },
    ],
  },
])

import { WidgetConfigProvider } from '@/contexts/WidgetConfigContext';

export default function App() {
  return (
    <NetworkProvider>
      <RouterProvider router={router} />
    </NetworkProvider>
  );
}

