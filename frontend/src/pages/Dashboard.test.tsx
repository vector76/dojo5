import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { Layout } from '../components/Layout'
import { Dashboard } from './Dashboard'
import { setToken } from '../api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

function renderDashboard(balanceData = { user_id: 5, expected_balance: 100, total_paid: 60, balance: 40 }) {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch')
    // GET /api/auth/me
    .mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 5, name: 'Test User', email: 'test@example.com', phone: '555-0001', role: 'user' }), { status: 200 }),
    )
    // GET /api/users/5/balance
    .mockResolvedValueOnce(
      new Response(JSON.stringify(balanceData), { status: 200 }),
    )
  return render(
    <MemoryRouter initialEntries={['/']}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route index element={<Dashboard />} />
          </Route>
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('Dashboard', () => {
  it('shows welcome message and balance summary', async () => {
    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText('Welcome, Test User!')).toBeInTheDocument()
    })

    await waitFor(() => {
      expect(screen.getByLabelText('Balance Summary')).toBeInTheDocument()
    })
    expect(screen.getByText('$100.00')).toBeInTheDocument()
    expect(screen.getByText('$60.00')).toBeInTheDocument()
    expect(screen.getByText('$40.00')).toBeInTheDocument()
    expect(screen.getByTestId('balance-status')).toHaveTextContent('Outstanding')
  })

  it('shows "Paid in full" for zero balance', async () => {
    renderDashboard({ user_id: 5, expected_balance: 50, total_paid: 50, balance: 0 })

    await waitFor(() => {
      expect(screen.getByTestId('balance-status')).toHaveTextContent('Paid in full')
    })
  })

  it('shows "Overpaid" for negative balance', async () => {
    renderDashboard({ user_id: 5, expected_balance: 50, total_paid: 70, balance: -20 })

    await waitFor(() => {
      expect(screen.getByTestId('balance-status')).toHaveTextContent('Overpaid')
    })
  })

  it('still renders dashboard when balance fetch fails', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ id: 5, name: 'Test User', email: 'test@example.com', phone: '555-0001', role: 'user' }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response('error', { status: 500 }),
      )

    render(
      <MemoryRouter initialEntries={['/']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route index element={<Dashboard />} />
            </Route>
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByText('Welcome, Test User!')).toBeInTheDocument()
    })
    // Balance should not be shown
    expect(screen.queryByLabelText('Balance Summary')).not.toBeInTheDocument()
  })
})
