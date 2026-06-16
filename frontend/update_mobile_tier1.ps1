cd C:\Users\USER\Downloads\noant\frontend

# ============================================
# 1. tailwind.config.ts
# ============================================
$tailwind = @'
import type { Config } from 'tailwindcss';

const config: Config = {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        noant: {
          black: '#0a0a0a',
          ink: '#171717',
          paper: '#fafafa',
          sky: '#0ea5e9',
          'sky-deep': '#0284c7',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
      },
      animation: {
        'spin-slow': 'spin 8s linear infinite',
        'fade-in': 'fadeIn 0.2s ease-out forwards',
        'slide-up': 'slideUp 0.3s ease-out forwards',
        shimmer: 'shimmer 1.5s infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(8px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        shimmer: {
          '0%': { backgroundPosition: '200% 0' },
          '100%': { backgroundPosition: '-200% 0' },
        },
      },
      boxShadow: {
        sky: '0 4px 14px 0 rgba(14, 165, 233, 0.39)',
      },
      transitionTimingFunction: {
        'ease-out-custom': 'cubic-bezier(0.16, 1, 0.3, 1)',
        'ease-in-out-custom': 'cubic-bezier(0.65, 0, 0.35, 1)',
      },
    },
  },
  plugins: [],
};

export default config;
'@
Set-Content -Path "tailwind.config.ts" -Value $tailwind -Encoding UTF8
Write-Host "✓ tailwind.config.ts"

# ============================================
# 2. src/index.css
# ============================================
$indexcss = @'
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --bg-base: #fafafa;
    --bg-surface: #ffffff;
    --bg-surface-hover: #f8fafc;
    --bg-elevated: #ffffff;
    --bg-inset: #f1f5f9;
    --bg-overlay: rgba(0, 0, 0, 0.5);
    --bg-chat-customer: #0a0a0a;
    
    --text-primary: #0a0a0a;
    --text-secondary: #64748b;
    --text-tertiary: #94a3b8;
    --text-inverse: #ffffff;
    --text-brand: #0ea5e9;
    --text-on-dark: #e5e5e5;
    
    --border-subtle: #f1f5f9;
    --border-default: #e2e8f0;
    --border-strong: #cbd5e1;
    
    --brand-sky: #0ea5e9;
    --brand-sky-deep: #0284c7;
    --brand-sky-soft: #e0f2fe;
    
    --success: #10b981;
    --success-soft: #d1fae5;
    --warning: #f59e0b;
    --warning-soft: #fef3c7;
    --error: #ef4444;
    --error-soft: #fee2e2;
    
    --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
    --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
    --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
    --shadow-sky: 0 4px 14px 0 rgba(14, 165, 233, 0.39);
    
    --ease-out: cubic-bezier(0.16, 1, 0.3, 1);
    --ease-in-out: cubic-bezier(0.65, 0, 0.35, 1);
  }

  @media (prefers-color-scheme: dark) {
    :root:not(.light):not(.dark) {
      --bg-base: #000000;
      --bg-surface: #0d0d0d;
      --bg-surface-hover: #1a1a1a;
      --bg-elevated: #141414;
      --bg-inset: #0a0a0a;
      --bg-overlay: rgba(0, 0, 0, 0.85);
      --bg-chat-customer: #1a1a1a;
      
      --text-primary: #e5e5e5;
      --text-secondary: #a1a1aa;
      --text-tertiary: #525252;
      --text-inverse: #0a0a0a;
      --text-brand: #38bdf8;
      --text-on-dark: #e5e5e5;
      
      --border-subtle: #1a1a1a;
      --border-default: #262626;
      --border-strong: #404040;
      
      --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.5);
      --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.6), 0 2px 4px -2px rgb(0 0 0 / 0.6);
      --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.7), 0 4px 6px -4px rgb(0 0 0 / 0.7);
    }
  }

  html.dark {
    --bg-base: #000000;
    --bg-surface: #0d0d0d;
    --bg-surface-hover: #1a1a1a;
    --bg-elevated: #141414;
    --bg-inset: #0a0a0a;
    --bg-overlay: rgba(0, 0, 0, 0.85);
    --bg-chat-customer: #1a1a1a;
    
    --text-primary: #e5e5e5;
    --text-secondary: #a1a1aa;
    --text-tertiary: #525252;
    --text-inverse: #0a0a0a;
    --text-brand: #38bdf8;
    --text-on-dark: #e5e5e5;
    
    --border-subtle: #1a1a1a;
    --border-default: #262626;
    --border-strong: #404040;
    
    --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.5);
    --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.6), 0 2px 4px -2px rgb(0 0 0 / 0.6);
    --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.7), 0 4px 6px -4px rgb(0 0 0 / 0.7);
  }

  html.light {
    --bg-base: #fafafa;
    --bg-surface: #ffffff;
    --bg-surface-hover: #f8fafc;
    --bg-elevated: #ffffff;
    --bg-inset: #f1f5f9;
    --bg-overlay: rgba(0, 0, 0, 0.5);
    --bg-chat-customer: #0a0a0a;
    
    --text-primary: #0a0a0a;
    --text-secondary: #64748b;
    --text-tertiary: #94a3b8;
    --text-inverse: #ffffff;
    --text-brand: #0ea5e9;
    --text-on-dark: #e5e5e5;
    
    --border-subtle: #f1f5f9;
    --border-default: #e2e8f0;
    --border-strong: #cbd5e1;
    
    --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
    --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
    --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
  }

  html {
    font-family: 'Inter', system-ui, -apple-system, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    background: var(--bg-base);
    color: var(--text-primary);
    -webkit-tap-highlight-color: transparent;
    touch-action: manipulation;
  }

  body {
    background: var(--bg-base);
    color: var(--text-primary);
    overscroll-behavior-y: none;
  }

  * {
    border-color: var(--border-default);
  }

  :focus-visible {
    outline: 2px solid var(--brand-sky);
    outline-offset: 2px;
  }

  @media (max-width: 1023px) {
    ::-webkit-scrollbar { display: none; }
    body { scrollbar-width: none; }
  }

  ::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }
  ::-webkit-scrollbar-track {
    background: transparent;
  }
  ::-webkit-scrollbar-thumb {
    background: var(--border-strong);
    border-radius: 3px;
  }
  ::-webkit-scrollbar-thumb:hover {
    background: var(--text-tertiary);
  }

  @media (prefers-reduced-motion: reduce) {
    *, *::before, *::after {
      animation-duration: 0.01ms !important;
      animation-iteration-count: 1 !important;
      transition-duration: 0.01ms !important;
      scroll-behavior: auto !important;
    }
  }
}

@layer utilities {
  .bg-surface { background: var(--bg-surface); }
  .bg-surface-hover:hover { background: var(--bg-surface-hover); }
  .bg-elevated { background: var(--bg-elevated); }
  .bg-inset { background: var(--bg-inset); }
  .text-primary { color: var(--text-primary); }
  .text-secondary { color: var(--text-secondary); }
  .text-tertiary { color: var(--text-tertiary); }
  .text-inverse { color: var(--text-inverse); }
  .border-subtle { border-color: var(--border-subtle); }
  .border-default { border-color: var(--border-default); }
  .border-strong { border-color: var(--border-strong); }
  .shadow-sky { box-shadow: var(--shadow-sky); }

  .animate-fade-in {
    animation: fadeIn 0.35s var(--ease-out) forwards;
  }
  .animate-slide-up {
    animation: slideUp 0.4s var(--ease-out) forwards;
  }
  .animate-shimmer {
    animation: shimmer 2s infinite;
    background: linear-gradient(90deg, var(--bg-inset) 25%, var(--bg-surface-hover) 50%, var(--bg-inset) 75%);
    background-size: 200% 100%;
  }

  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border-width: 0;
  }

  .scrollbar-thin {
    scrollbar-width: thin;
    scrollbar-color: var(--border-strong) transparent;
  }

  .pb-safe { padding-bottom: env(safe-area-inset-bottom); }
  .pt-safe { padding-top: env(safe-area-inset-top); }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  @keyframes slideUp {
    from { opacity: 0; transform: translateY(12px); }
    to { opacity: 1; transform: translateY(0); }
  }
  @keyframes shimmer {
    0% { background-position: 200% 0; }
    100% { background-position: -200% 0; }
  }
}
'@
Set-Content -Path "src\index.css" -Value $indexcss -Encoding UTF8
Write-Host "✓ src/index.css"

# ============================================
# 3. src/app/(dashboard)/layout.tsx
# ============================================
$dashLayout = @'
import { Outlet } from 'react-router-dom'
import { DashboardLayout } from '@/components/layout/DashboardLayout'

export default function DashboardRouteLayout() {
  return (
    <DashboardLayout>
      <Outlet />
    </DashboardLayout>
  )
}
'@
Set-Content -Path "src\app\(dashboard)\layout.tsx" -Value $dashLayout -Encoding UTF8
Write-Host "✓ src/app/(dashboard)/layout.tsx"

# ============================================
# 4. src/components/layout/DashboardLayout.tsx
# ============================================
$compLayout = @'
import { useState, ReactNode } from 'react'
import { cn } from '@/lib/utils'
import { Sidebar } from './Sidebar'
import { Header } from './Header'
import { BottomNav } from './BottomNav'
import { MobileOverlay } from './MobileOverlay'

interface DashboardLayoutProps {
  children: ReactNode
}

export function DashboardLayout({ children }: DashboardLayoutProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false)

  return (
    <div className="min-h-screen bg-base flex overflow-x-hidden">
      {/* Desktop sidebar */}
      <div className="hidden lg:block fixed top-0 left-0 w-[220px] h-screen z-40">
        <Sidebar />
      </div>

      {/* Mobile sidebar */}
      <div
        className={cn(
          'fixed top-0 left-0 w-[280px] h-screen z-50 lg:hidden transition-transform duration-300 ease-[cubic-bezier(0.16,1,0.3,1)]',
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        )}
      >
        <Sidebar onClose={() => setSidebarOpen(false)} />
      </div>

      <MobileOverlay open={sidebarOpen} onClose={() => setSidebarOpen(false)} />

      {/* Main content */}
      <div className="flex-1 lg:ml-[220px] flex flex-col min-h-screen pb-[calc(3.5rem+env(safe-area-inset-bottom))] lg:pb-0">
        <Header onMenuClick={() => setSidebarOpen(true)} />
        <main className="flex-1 p-4 lg:p-6 overflow-y-auto bg-base overscroll-contain">
          {children}
        </main>
        <BottomNav />
      </div>
    </div>
  )
}
'@
Set-Content -Path "src\components\layout\DashboardLayout.tsx" -Value $compLayout -Encoding UTF8
Write-Host "✓ src/components/layout/DashboardLayout.tsx"

# ============================================
# 5. src/components/layout/Sidebar.tsx
# ============================================
$sidebar = @'
import { cn } from '@/lib/utils'
import { NavLink } from 'react-router-dom'
import { LayoutGrid, MessageSquare, GraduationCap, BarChart3, Link2, Settings, LogOut, X } from 'lucide-react'
import { useAuth } from '@/hooks/useAuth'

const navSections = [
  {
    title: 'Your workspace',
    items: [
      { to: '/', icon: LayoutGrid, label: 'Overview' },
      { to: '/chats', icon: MessageSquare, label: 'Conversations' },
      { to: '/insights', icon: BarChart3, label: 'Insights' },
    ],
  },
  {
    title: 'Build your Noant',
    items: [
      { to: '/teach', icon: GraduationCap, label: 'Teach your Noant' },
      { to: '/channels', icon: Link2, label: 'Your channels' },
    ],
  },
  {
    title: 'Manage',
    items: [
      { to: '/setup', icon: Settings, label: 'Your setup' },
    ],
  },
]

interface SidebarProps {
  onClose?: () => void
}

export function Sidebar({ onClose }: SidebarProps) {
  const { user, signOut } = useAuth()

  const initials = user
    ? `${user.first_name[0]}${user.last_name[0]}`.toUpperCase()
    : '--'

  const name = user
    ? `${user.first_name} ${user.last_name}`
    : 'Loading...'

  return (
    <div className="flex flex-col h-full bg-surface border-r border-default">
      <div className="flex items-center justify-between p-4 border-b border-default">
        <div className="flex items-center gap-2.5">
          <svg className="w-7 h-7 animate-spin-slow" viewBox="0 0 100 100" fill="none">
            <circle cx="50" cy="50" r="42" stroke="currentColor" strokeWidth="6" strokeDasharray="8 10" strokeLinecap="round" className="text-primary" />
            <circle cx="50" cy="50" r="28" fill="currentColor" className="text-primary" />
            <circle cx="40" cy="50" r="3" fill="var(--bg-base)" />
            <circle cx="50" cy="50" r="4" fill="var(--bg-base)" />
            <circle cx="60" cy="50" r="5" fill="var(--bg-base)" />
          </svg>
          <span className="text-lg font-bold tracking-widest lowercase text-primary">noant</span>
        </div>
        {onClose && (
          <button
            onClick={onClose}
            className="lg:hidden w-8 h-8 rounded-full bg-inset flex items-center justify-center text-secondary hover:text-primary active:scale-95 transition-all"
            aria-label="Close menu"
          >
            <X className="w-4 h-4" />
          </button>
        )}
      </div>

      <nav className="flex-1 p-3 overflow-y-auto">
        {navSections.map((section) => (
          <div key={section.title} className="mb-5">
            <div className="px-3 mb-2 text-[10px] font-semibold uppercase tracking-widest text-tertiary">
              {section.title}
            </div>
            {section.items.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.to === '/'}
                onClick={() => onClose?.()}
                className={({ isActive }) =>
                  cn(
                    'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 group active:scale-[0.98]',
                    isActive
                      ? 'bg-noant-sky/10 text-noant-sky-deep dark:text-noant-sky border-r-2 border-noant-sky'
                      : 'text-secondary hover:bg-inset hover:text-primary'
                  )
                }
              >
                <item.icon className="w-[18px] h-[18px] group-hover:text-noant-sky transition-colors" strokeWidth={2} />
                <span>{item.label}</span>
              </NavLink>
            ))}
          </div>
        ))}
      </nav>

      <div className="p-3 border-t border-default">
        <button
          onClick={signOut}
          className="flex items-center gap-3 w-full px-3 py-2.5 rounded-lg hover:bg-inset transition-colors active:scale-[0.98]"
        >
          <div className="w-8 h-8 rounded-full bg-noant-black text-white dark:bg-white dark:text-noant-black flex items-center justify-center text-xs font-semibold">
            {initials}
          </div>
          <div className="flex-1 min-w-0 text-left">
            <div className="text-sm font-semibold truncate text-primary">{name}</div>
            <div className="text-[10px] text-tertiary">{user?.plan || 'Free'} plan</div>
          </div>
          <LogOut className="w-4 h-4 text-tertiary hover:text-noant-sky transition-colors" />
        </button>
      </div>
    </div>
  )
}
'@
Set-Content -Path "src\components\layout\Sidebar.tsx" -Value $sidebar -Encoding UTF8
Write-Host "✓ src/components/layout/Sidebar.tsx"

# ============================================
# 6. src/components/layout/Header.tsx
# ============================================
$header = @'
import { useLocation } from 'react-router-dom'
import { Sun, Moon, Bell, Menu } from 'lucide-react'
import { cn } from '@/lib/utils'

const titles: Record<string, string> = {
  '/': 'Overview',
  '/chats': 'Conversations',
  '/teach': 'Teach your Noant',
  '/insights': 'Insights',
  '/channels': 'Your channels',
  '/setup': 'Your setup',
}

export function Header({ onMenuClick }: { onMenuClick: () => void }) {
  const location = useLocation()
  const title = titles[location.pathname] || 'noant'

  const toggleTheme = () => {
    const html = document.documentElement
    const isDark = html.classList.contains('dark')
    
    if (isDark) {
      html.classList.remove('dark')
      html.classList.add('light')
      localStorage.setItem('noant_theme', 'light')
    } else {
      html.classList.remove('light')
      html.classList.add('dark')
      localStorage.setItem('noant_theme', 'dark')
    }
  }

  return (
    <header className="sticky top-0 z-40 h-12 lg:h-14 bg-surface/80 backdrop-blur-xl border-b border-default flex items-center justify-between px-3 lg:px-4">
      <div className="flex items-center gap-3">
        <button
          onClick={onMenuClick}
          className="lg:hidden w-9 h-9 rounded-full bg-inset flex items-center justify-center text-secondary hover:bg-surface-hover active:scale-95 transition-all"
          aria-label="Open menu"
        >
          <Menu className="w-[18px] h-[18px]" strokeWidth={2} />
        </button>
        <h1 className="text-sm lg:text-base font-semibold text-primary tracking-tight">{title}</h1>
      </div>
      <div className="flex items-center gap-1.5">
        <button
          onClick={toggleTheme}
          className="w-9 h-9 rounded-full bg-inset flex items-center justify-center text-secondary hover:bg-surface-hover active:scale-95 transition-all"
          aria-label="Toggle theme"
        >
          <Sun className="w-[18px] h-[18px] block dark:hidden" strokeWidth={2} />
          <Moon className="w-[18px] h-[18px] hidden dark:block" strokeWidth={2} />
        </button>
        <button className="w-9 h-9 rounded-full bg-inset flex items-center justify-center text-secondary hover:bg-surface-hover active:scale-95 transition-all relative">
          <Bell className="w-[18px] h-[18px]" strokeWidth={2} />
          <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full ring-2 ring-surface" />
        </button>
      </div>
    </header>
  )
}
'@
Set-Content -Path "src\components\layout\Header.tsx" -Value $header -Encoding UTF8
Write-Host "✓ src/components/layout/Header.tsx"

# ============================================
# 7. src/components/layout/BottomNav.tsx
# ============================================
$bottomNav = @'
import { NavLink } from 'react-router-dom'
import { LayoutGrid, MessageSquare, BarChart3, Link2, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'

const items = [
  { to: '/', icon: LayoutGrid, label: 'Home' },
  { to: '/chats', icon: MessageSquare, label: 'Inbox' },
  { to: '/insights', icon: BarChart3, label: 'Insights' },
  { to: '/channels', icon: Link2, label: 'Channels' },
  { to: '/setup', icon: Settings, label: 'Setup' },
]

export function BottomNav() {
  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 lg:hidden bg-surface/95 backdrop-blur-lg border-t border-default pb-[env(safe-area-inset-bottom)]">
      <div className="flex h-14">
        {items.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === '/'}
            className={({ isActive }) =>
              cn(
                'flex-1 flex flex-col items-center justify-center gap-0.5 relative select-none active:opacity-70 transition-colors duration-200',
                isActive ? 'text-noant-sky' : 'text-tertiary'
              )
            }
          >
            {({ isActive }) => (
              <>
                <div className="relative flex items-center justify-center w-7 h-7">
                  <item.icon
                    className={cn(
                      'w-[22px] h-[22px] transition-transform duration-200',
                      isActive && 'scale-110'
                    )}
                    strokeWidth={isActive ? 2.5 : 1.5}
                  />
                </div>
                <span className={cn(
                  'text-[10px] font-medium leading-none transition-all',
                  isActive && 'font-semibold'
                )}>
                  {item.label}
                </span>
                {isActive && (
                  <span className="absolute top-1 left-1/2 -translate-x-1/2 w-1 h-1 rounded-full bg-noant-sky" />
                )}
              </>
            )}
          </NavLink>
        ))}
      </div>
    </nav>
  )
}
'@
Set-Content -Path "src\components\layout\BottomNav.tsx" -Value $bottomNav -Encoding UTF8
Write-Host "✓ src/components/layout/BottomNav.tsx"

# ============================================
# 8. src/components/layout/MobileOverlay.tsx
# ============================================
$overlay = @'
import { cn } from '@/lib/utils'

export function MobileOverlay({ open, onClose }: { open: boolean; onClose: () => void }) {
  return (
    <div
      onClick={onClose}
      className={cn(
        'fixed inset-0 bg-black/40 backdrop-blur-sm z-40 lg:hidden transition-opacity duration-300',
        open ? 'opacity-100 pointer-events-auto' : 'opacity-0 pointer-events-none'
      )}
    />
  )
}
'@
Set-Content -Path "src\components\layout\MobileOverlay.tsx" -Value $overlay -Encoding UTF8
Write-Host "✓ src/components/layout/MobileOverlay.tsx"

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "Tier 1 mobile shell updated successfully!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "`nNext: Run this to dump Tier 2 (chat files):"
Write-Host 'Get-Content "src\app\(dashboard)\chats\page.tsx","src\components\chat\ChatList.tsx","src\components\chat\ChatMessages.tsx","src\components\chat\ChatInput.tsx","src\components\chat\CustomerInfo.tsx" -Raw' -ForegroundColor Cyan