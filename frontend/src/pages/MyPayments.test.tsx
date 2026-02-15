import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { Layout } from '../components/Layout'
import { MyPayments } from './MyPayments'
import { setToken } from '../api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

const mockPayments = [
  { id: 1, user_id: 5, amount: 50, date: '2025-06-01', note: 'Monthly fee', recorded_by: 1 },
  { id: 2, user_id: 5, amount: 30, date: '2025-07-01', note: null, recorded_by: 1 },
]

const mockBalance = {
  user_id: 5,
  expected_balance: 200,
  total_paid: 80,
  balance: 120,
}

function renderMyPayments() {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch')
    // GET /api/auth/me
    .mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 5, name: 'Regular User', email: 'user@example.com', phone: '555-5555', role: 'user' }), { status: 200 }),
    )
    // GET /api/users/5/payments
    .mockResolvedValueOnce(
      new Response(JSON.stringify(mockPayments), { status: 200 }),
    )
    // GET /api/users/5/balance
    .mockResolvedValueOnce(
      new Response(JSON.stringify(mockBalance), { status: 200 }),
    )
  return render(
    <MemoryRouter initialEntries={['/my-payments']}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route path="my-payments" element={<MyPayments />} />
          </Route>
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('MyPayments', () => {
  it('renders payment history with balance', async () => {
    renderMyPayments()

    await waitFor(() => {
      expect(screen.getByText('My Payments')).toBeInTheDocument()
    })

    await waitFor(() => {
      expect(screen.getByText('$200.00')).toBeInTheDocument()
    })
    expect(screen.getByText('$80.00')).toBeInTheDocument()
    expect(screen.getByText('$120.00')).toBeInTheDocument()
  })

  it('renders payment table', async () => {
    renderMyPayments()

    await waitFor(() => {
      expect(screen.getByText('$50.00')).toBeInTheDocument()
    })
    expect(screen.getByText('$30.00')).toBeInTheDocument()
    expect(screen.getByText('2025-06-01')).toBeInTheDocument()
    expect(screen.getByText('2025-07-01')).toBeInTheDocument()
    expect(screen.getByText('Monthly fee')).toBeInTheDocument()
  })

  it('shows "No payments recorded" when empty', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ id: 5, name: 'User', email: 'user@example.com', phone: '555', role: 'user' }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify([]), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ user_id: 5, expected_balance: 0, total_paid: 0, balance: 0 }), { status: 200 }),
      )

    render(
      <MemoryRouter initialEntries={['/my-payments']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="my-payments" element={<MyPayments />} />
            </Route>
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByText('No payments recorded')).toBeInTheDocument()
    })
  })

  it('shows error on API failure', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ id: 5, name: 'User', email: 'user@example.com', phone: '555', role: 'user' }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response('server error', { status: 500 }),
      )

    render(
      <MemoryRouter initialEntries={['/my-payments']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="my-payments" element={<MyPayments />} />
            </Route>
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to load payment history')
    })
  })

  it('shows My Payments link in nav for all users', async () => {
    renderMyPayments()

    await waitFor(() => {
      expect(screen.getByText('My Payments')).toBeInTheDocument()
    })

    const nav = screen.getByRole('navigation')
    expect(nav.querySelector('a[href="/my-payments"]')).toBeInTheDocument()
  })
})
