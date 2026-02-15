import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { Layout } from '../components/Layout'
import { Payments } from './Payments'
import { setToken } from '../api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

const mockMembers = [
  { id: 1, name: 'Alice Admin', email: 'alice@example.com', phone: '555-0001', role: 'admin' },
  { id: 2, name: 'Bob User', email: 'bob@example.com', phone: '555-0002', role: 'user' },
  { id: 3, name: 'Carol User', email: 'carol@example.com', phone: '555-0003', role: 'user' },
]

const mockPayments = [
  { id: 1, user_id: 2, amount: 50, date: '2025-06-01', note: 'Monthly fee', recorded_by: 1 },
  { id: 2, user_id: 2, amount: 25, date: '2025-07-01', note: null, recorded_by: 1 },
]

const mockBalance = {
  user_id: 2,
  expected_balance: 100,
  total_paid: 75,
  balance: 25,
}

function renderPayments() {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch')
    // GET /api/auth/me
    .mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 1, name: 'Admin', email: 'admin@example.com', phone: '555-0001', role: 'admin' }), { status: 200 }),
    )
    // GET /api/users
    .mockResolvedValueOnce(
      new Response(JSON.stringify(mockMembers), { status: 200 }),
    )
  return render(
    <MemoryRouter initialEntries={['/payments']}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route path="payments" element={<Payments />} />
          </Route>
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

function getRecordSection() {
  return within(screen.getByRole('region', { name: 'Record Payment' }))
}

function getHistorySection() {
  return within(screen.getByRole('region', { name: 'Payment History' }))
}

describe('Payments', () => {
  it('renders payment page with record form and history section', async () => {
    renderPayments()

    await waitFor(() => {
      expect(screen.getByRole('region', { name: 'Record Payment' })).toBeInTheDocument()
    })
    expect(screen.getByRole('region', { name: 'Payment History' })).toBeInTheDocument()
    expect(screen.getByLabelText('Amount')).toBeInTheDocument()
    expect(screen.getByLabelText('Date')).toBeInTheDocument()
    expect(screen.getByLabelText('Note')).toBeInTheDocument()
  })

  it('populates member dropdown', async () => {
    renderPayments()

    await waitFor(() => {
      expect(screen.getByRole('region', { name: 'Record Payment' })).toBeInTheDocument()
    })

    const memberSelect = document.getElementById('paymentMember') as HTMLSelectElement
    // Should have 3 members + "Select a member" placeholder
    expect(memberSelect.options).toHaveLength(4)
  })

  it('records a payment', async () => {
    const user = userEvent.setup()
    renderPayments()

    await waitFor(() => {
      expect(screen.getByRole('region', { name: 'Record Payment' })).toBeInTheDocument()
    })

    const section = getRecordSection()
    await user.selectOptions(section.getByLabelText('Member'), '2')
    await user.type(screen.getByLabelText('Amount'), '50')
    await user.type(screen.getByLabelText('Note'), 'Monthly fee')

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 3, user_id: 2, amount: 50, date: '2025-06-15', note: 'Monthly fee', recorded_by: 1 }), { status: 201 }),
    )

    await user.click(screen.getByRole('button', { name: 'Record Payment' }))

    await waitFor(() => {
      expect(screen.getByText('Payment recorded')).toBeInTheDocument()
    })
  })

  it('shows error on record failure', async () => {
    const user = userEvent.setup()
    renderPayments()

    await waitFor(() => {
      expect(screen.getByRole('region', { name: 'Record Payment' })).toBeInTheDocument()
    })

    const section = getRecordSection()
    await user.selectOptions(section.getByLabelText('Member'), '2')
    await user.type(screen.getByLabelText('Amount'), '50')

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response('server error', { status: 500 }),
    )

    await user.click(screen.getByRole('button', { name: 'Record Payment' }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to record payment')
    })
  })

  it('views payment history for a member', async () => {
    const user = userEvent.setup()
    renderPayments()

    await waitFor(() => {
      expect(screen.getByRole('region', { name: 'Payment History' })).toBeInTheDocument()
    })

    const section = getHistorySection()
    await user.selectOptions(section.getByLabelText('Member'), '2')

    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify(mockPayments), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(mockBalance), { status: 200 }),
      )

    await user.click(screen.getByRole('button', { name: 'View History' }))

    await waitFor(() => {
      expect(section.getByText('Monthly fee')).toBeInTheDocument()
    })
    expect(section.getByText('$100.00')).toBeInTheDocument() // expected balance
    expect(section.getByText('$75.00')).toBeInTheDocument() // total paid
  })

  it('shows "No payments recorded" for member with no payments', async () => {
    const user = userEvent.setup()
    renderPayments()

    await waitFor(() => {
      expect(screen.getByRole('region', { name: 'Payment History' })).toBeInTheDocument()
    })

    const section = getHistorySection()
    await user.selectOptions(section.getByLabelText('Member'), '3')

    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify([]), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ user_id: 3, expected_balance: 0, total_paid: 0, balance: 0 }), { status: 200 }),
      )

    await user.click(screen.getByRole('button', { name: 'View History' }))

    await waitFor(() => {
      expect(section.getByText('No payments recorded')).toBeInTheDocument()
    })
  })

  it('shows error on load failure', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ id: 1, name: 'Admin', email: 'admin@example.com', phone: '555-0001', role: 'admin' }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response('server error', { status: 500 }),
      )

    render(
      <MemoryRouter initialEntries={['/payments']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="payments" element={<Payments />} />
            </Route>
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to load members')
    })
  })
})
