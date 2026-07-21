import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { AuthLayout } from '@/components/layout/AuthLayout'

function renderWithRouter(initialEntries = ['/login']) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <Routes>
        <Route element={<AuthLayout />}>
          <Route path="/login" element={<div>child content</div>} />
        </Route>
      </Routes>
    </MemoryRouter>
  )
}

describe('AuthLayout', () => {
  it('renders the layout container', () => {
    renderWithRouter()
  })

  it('renders the brand panel with noant logo', () => {
    const { container } = renderWithRouter()
    expect(container.querySelector('.bg-noant-black')).toBeInTheDocument()
  })

  it('renders the quote text', () => {
    renderWithRouter()
    expect(screen.getByText(/I used to wake up to 47 unread messages/)).toBeInTheDocument()
  })

  it('renders stats: 500+ Languages', () => {
    renderWithRouter()
    expect(screen.getByText('500+')).toBeInTheDocument()
    expect(screen.getByText('Languages')).toBeInTheDocument()
  })

  it('renders stats: 24/7 Always on', () => {
    renderWithRouter()
    expect(screen.getByText('24/7')).toBeInTheDocument()
    expect(screen.getByText('Always on')).toBeInTheDocument()
  })

  it('renders stats: <2s Response time', () => {
    renderWithRouter()
    expect(screen.getByText('<2s')).toBeInTheDocument()
    expect(screen.getByText('Response time')).toBeInTheDocument()
  })

  it('renders the noant brand name', () => {
    renderWithRouter()
    expect(screen.getByText('noant')).toBeInTheDocument()
  })

  it('renders child content via Outlet', () => {
    renderWithRouter()
    expect(screen.getByText('child content')).toBeInTheDocument()
  })
})
