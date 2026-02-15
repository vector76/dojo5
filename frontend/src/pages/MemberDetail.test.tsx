import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { Layout } from '../components/Layout'
import { MemberDetail } from './MemberDetail'
import { setToken } from '../api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

const memberData = {
  id: 2,
  name: 'Bob Instructor',
  email: 'bob@example.com',
  phone: '555-0002',
  role: 'instructor',
  membership_type: 'monthly',
  membership_status: 'active',
  emergency_contact: 'Jane 555-9999',
  join_date: '2025-01-15',
}

const balanceData = {
  user_id: 2,
  expected_balance: 100,
  total_paid: 60,
  balance: 40,
}

function mockFetchForAdmin() {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch')
    .mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 1, name: 'Admin', email: 'admin@example.com', phone: '555-0001', role: 'admin' }), { status: 200 }),
    )
    .mockResolvedValueOnce(
      new Response(JSON.stringify(memberData), { status: 200 }),
    )
    .mockResolvedValueOnce(
      new Response(JSON.stringify(balanceData), { status: 200 }),
    )
}

function mockFetchForInstructor() {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch')
    .mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 3, name: 'Instructor', email: 'inst@example.com', phone: '555-0003', role: 'instructor' }), { status: 200 }),
    )
    .mockResolvedValueOnce(
      new Response(JSON.stringify(memberData), { status: 200 }),
    )
}

function renderMemberDetail(mockFn = mockFetchForAdmin) {
  mockFn()
  return render(
    <MemoryRouter initialEntries={['/members/2']}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route path="members/:id" element={<MemberDetail />} />
            <Route path="members" element={<div>Members List</div>} />
          </Route>
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('MemberDetail', () => {
  it('renders member details', async () => {
    renderMemberDetail()

    await waitFor(() => {
      expect(screen.getByText('Bob Instructor')).toBeInTheDocument()
    })
    expect(screen.getByText('bob@example.com')).toBeInTheDocument()
    expect(screen.getByText('555-0002')).toBeInTheDocument()
    expect(screen.getByText('instructor')).toBeInTheDocument()
    expect(screen.getByText('monthly')).toBeInTheDocument()
    expect(screen.getByText('active')).toBeInTheDocument()
    expect(screen.getByText('Jane 555-9999')).toBeInTheDocument()
    expect(screen.getByText('2025-01-15')).toBeInTheDocument()
  })

  it('shows balance info for admin', async () => {
    renderMemberDetail()

    await waitFor(() => {
      expect(screen.getByText('$100.00')).toBeInTheDocument()
    })
    expect(screen.getByText('$60.00')).toBeInTheDocument()
    expect(screen.getByText('$40.00')).toBeInTheDocument()
  })

  it('hides admin-only sections for non-admin', async () => {
    renderMemberDetail(mockFetchForInstructor)

    await waitFor(() => {
      expect(screen.getByText('Bob Instructor')).toBeInTheDocument()
    })
    expect(screen.queryByText('Reset Password')).not.toBeInTheDocument()
    expect(screen.queryByText('Balance')).not.toBeInTheDocument()
    expect(screen.queryByText('Delete Member')).not.toBeInTheDocument()
  })

  it('switches to edit mode and saves changes', async () => {
    const user = userEvent.setup()
    renderMemberDetail()

    await waitFor(() => {
      expect(screen.getByText('Bob Instructor')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Edit'))

    // Edit form should be visible
    expect(screen.getByLabelText('Name')).toHaveValue('Bob Instructor')
    expect(screen.getByLabelText('Email')).toHaveValue('bob@example.com')

    // Modify name
    await user.clear(screen.getByLabelText('Name'))
    await user.type(screen.getByLabelText('Name'), 'Bob Updated')

    // Mock the PUT response
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ ...memberData, name: 'Bob Updated' }), { status: 200 }),
    )

    await user.click(screen.getByText('Save'))

    await waitFor(() => {
      expect(screen.getByText('Bob Updated')).toBeInTheDocument()
    })
    // Should be back in view mode
    expect(screen.getByText('Edit')).toBeInTheDocument()
  })

  it('shows admin-only fields in edit mode for admin', async () => {
    const user = userEvent.setup()
    renderMemberDetail()

    await waitFor(() => {
      expect(screen.getByText('Bob Instructor')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Edit'))

    expect(screen.getByLabelText('Role')).toBeInTheDocument()
    expect(screen.getByLabelText('Membership Type')).toBeInTheDocument()
    expect(screen.getByLabelText('Membership Status')).toBeInTheDocument()
    expect(screen.getByLabelText('Join Date')).toBeInTheDocument()
  })

  it('hides admin-only fields in edit mode for non-admin', async () => {
    const user = userEvent.setup()
    renderMemberDetail(mockFetchForInstructor)

    await waitFor(() => {
      expect(screen.getByText('Bob Instructor')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Edit'))

    expect(screen.getByLabelText('Name')).toBeInTheDocument()
    expect(screen.queryByLabelText('Role')).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Membership Type')).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Membership Status')).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Join Date')).not.toBeInTheDocument()
  })

  it('handles password reset', async () => {
    const user = userEvent.setup()
    renderMemberDetail()

    await waitFor(() => {
      expect(screen.getByLabelText('New Password')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('New Password'), 'newpassword123')

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(null, { status: 204 }),
    )

    await user.click(screen.getByRole('button', { name: 'Reset Password' }))

    await waitFor(() => {
      expect(screen.getByText('Password updated')).toBeInTheDocument()
    })
  })

  it('handles balance update', async () => {
    const user = userEvent.setup()
    renderMemberDetail()

    await waitFor(() => {
      expect(screen.getByLabelText('Set Expected Balance')).toBeInTheDocument()
    })

    await user.clear(screen.getByLabelText('Set Expected Balance'))
    await user.type(screen.getByLabelText('Set Expected Balance'), '200')

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ user_id: 2, expected_balance: 200, total_paid: 60, balance: 140 }), { status: 200 }),
    )

    await user.click(screen.getByText('Update Balance'))

    await waitFor(() => {
      expect(screen.getByText('Balance updated')).toBeInTheDocument()
    })
    expect(screen.getByText('$200.00')).toBeInTheDocument()
  })

  it('shows error on API failure', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ id: 1, name: 'Admin', email: 'admin@example.com', phone: '555-0001', role: 'admin' }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response('not found', { status: 404 }),
      )

    render(
      <MemoryRouter initialEntries={['/members/999']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="members/:id" element={<MemberDetail />} />
            </Route>
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to load member')
    })
  })

  it('cancels edit mode and restores original values', async () => {
    const user = userEvent.setup()
    renderMemberDetail()

    await waitFor(() => {
      expect(screen.getByText('Bob Instructor')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Edit'))

    await user.clear(screen.getByLabelText('Name'))
    await user.type(screen.getByLabelText('Name'), 'Changed Name')

    await user.click(screen.getByText('Cancel'))

    // Should be back in view mode with original name
    expect(screen.getByText('Bob Instructor')).toBeInTheDocument()
    expect(screen.queryByText('Changed Name')).not.toBeInTheDocument()
  })
})
