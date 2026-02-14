import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { ProtectedRoute } from './ProtectedRoute'
import { Layout } from './Layout'
import { setToken } from '../api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

function mockUser(role: string, name = 'Test User') {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch').mockResolvedValue(
    new Response(
      JSON.stringify({
        id: 1,
        name,
        email: 'test@example.com',
        phone: '555-0001',
        role,
      }),
      { status: 200 },
    ),
  )
}

function renderLayout() {
  return render(
    <MemoryRouter initialEntries={['/']}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route index element={<div>Dashboard Page</div>} />
          </Route>
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('Layout', () => {
  it('shows all nav links for admin', async () => {
    mockUser('admin')
    renderLayout()

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument()
    })
    expect(screen.getByText('Classes')).toBeInTheDocument()
    expect(screen.getByText('Members')).toBeInTheDocument()
    expect(screen.getByText('Payments')).toBeInTheDocument()
    expect(screen.getByText('Class Types')).toBeInTheDocument()
  })

  it('shows Members but not Payments/Class Types for instructor', async () => {
    mockUser('instructor')
    renderLayout()

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument()
    })
    expect(screen.getByText('Classes')).toBeInTheDocument()
    expect(screen.getByText('Members')).toBeInTheDocument()
    expect(screen.queryByText('Payments')).not.toBeInTheDocument()
    expect(screen.queryByText('Class Types')).not.toBeInTheDocument()
  })

  it('shows only Dashboard and Classes for regular user', async () => {
    mockUser('user')
    renderLayout()

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument()
    })
    expect(screen.getByText('Classes')).toBeInTheDocument()
    expect(screen.queryByText('Members')).not.toBeInTheDocument()
    expect(screen.queryByText('Payments')).not.toBeInTheDocument()
    expect(screen.queryByText('Class Types')).not.toBeInTheDocument()
  })

  it('displays user name and logout button', async () => {
    mockUser('admin', 'Jane Admin')
    renderLayout()

    await waitFor(() => {
      expect(screen.getByText('Jane Admin')).toBeInTheDocument()
    })
    expect(screen.getByText('Logout')).toBeInTheDocument()
  })

  it('renders the outlet content', async () => {
    mockUser('admin')
    renderLayout()

    await waitFor(() => {
      expect(screen.getByText('Dashboard Page')).toBeInTheDocument()
    })
  })
})
