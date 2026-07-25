import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Topbar } from './Topbar'
import { CommandPalette } from './CommandPalette'

export function Shell() {
  return (
    <div className="min-h-screen bg-bg-base">
      <CommandPalette />
      <Sidebar />
      <div className="pl-60">
        <Topbar />
        <main className="p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
