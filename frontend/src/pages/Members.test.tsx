import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { Layout } from '../components/Layout'
import { Members } from './Members'
import { setToken } from '../api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

const mockMembers = [
  {
    id: 1,
    name: 'Alice Admin',
    email: 'alice@example.com',
    phone: '555-0001',
    role: 'admin',
    membership_status: 'active',
  },
  {
    id: 2,
    name: 'Bob Instructor',
    email: 'bob@example.com',
    phone: '555-0002',
    role: 'instructor',
    membership_status: 'active',
  },
  {
    id: 3,
    name: 'Carol User',
    email: 'carol@example.com',
    phone: '555-0003',
    role: 'user',
    membership_status: 'inactive',
  },
  {
    id: 4,
    name: 'Dave User',
    email: 'dave@example.com',
    phone: '555-0004',
    role: 'user',
    membership_status: 'active',
  },
]

function mockAuthAndMembers(role: string) {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch')
    // GET /api/auth/me
    .mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          id: 1,
          name: 'Test User',
          email: 'test@example.com',
          phone: '555-0001',
          role,
        }),
        { status: 200 },
      ),
    )
    // GET /api/users
    .mockResolvedValueOnce(
      new Response(JSON.stringify(mockMembers), { status: 200 }),
    )
}

function renderMembers(role = 'admin') {
  mockAuthAndMembers(role)
  return render(
    <MemoryRouter initialEntries={['/members']}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route path="/members" element={<Members />} />
          </Route>
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('Members', () => {
  it('renders member list', async () => {
    renderMembers()

    await waitFor(() => {
      expect(screen.getByText('Alice Admin')).toBeInTheDocument()
    })
    expect(screen.getByText('Bob Instructor')).toBeInTheDocument()
    expect(screen.getByText('Carol User')).toBeInTheDocument()
    expect(screen.getByText('Dave User')).toBeInTheDocument()
  })

  it('filters by search text (name)', async () => {
    const user = userEvent.setup()
    renderMembers()

    await waitFor(() => {
      expect(screen.getByText('Alice Admin')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('Search members'), 'carol')

    expect(screen.getByText('Carol User')).toBeInTheDocument()
    expect(screen.queryByText('Alice Admin')).not.toBeInTheDocument()
    expect(screen.queryByText('Bob Instructor')).not.toBeInTheDocument()
    expect(screen.queryByText('Dave User')).not.toBeInTheDocument()
  })

  it('filters by search text (email)', async () => {
    const user = userEvent.setup()
    renderMembers()

    await waitFor(() => {
      expect(screen.getByText('Alice Admin')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('Search members'), 'bob@')

    expect(screen.getByText('Bob Instructor')).toBeInTheDocument()
    expect(screen.queryByText('Alice Admin')).not.toBeInTheDocument()
  })

  it('filters by role', async () => {
    const user = userEvent.setup()
    renderMembers()

    await waitFor(() => {
      expect(screen.getByText('Alice Admin')).toBeInTheDocument()
    })

    await user.selectOptions(screen.getByLabelText('Filter by role'), 'user')

    expect(screen.getByText('Carol User')).toBeInTheDocument()
    expect(screen.getByText('Dave User')).toBeInTheDocument()
    expect(screen.queryByText('Alice Admin')).not.toBeInTheDocument()
    expect(screen.queryByText('Bob Instructor')).not.toBeInTheDocument()
  })

  it('filters by membership status', async () => {
    const user = userEvent.setup()
    renderMembers()

    await waitFor(() => {
      expect(screen.getByText('Alice Admin')).toBeInTheDocument()
    })

    await user.selectOptions(screen.getByLabelText('Filter by status'), 'inactive')

    expect(screen.getByText('Carol User')).toBeInTheDocument()
    expect(screen.queryByText('Alice Admin')).not.toBeInTheDocument()
    expect(screen.queryByText('Dave User')).not.toBeInTheDocument()
  })

  it('shows "No members found" when filters match nothing', async () => {
    const user = userEvent.setup()
    renderMembers()

    await waitFor(() => {
      expect(screen.getByText('Alice Admin')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('Search members'), 'nonexistent')

    expect(screen.getByText('No members found')).toBeInTheDocument()
  })

  it('combines search and role filter', async () => {
    const user = userEvent.setup()
    renderMembers()

    await waitFor(() => {
      expect(screen.getByText('Alice Admin')).toBeInTheDocument()
    })

    await user.selectOptions(screen.getByLabelText('Filter by role'), 'user')
    await user.type(screen.getByLabelText('Search members'), 'dave')

    expect(screen.getByText('Dave User')).toBeInTheDocument()
    expect(screen.queryByText('Carol User')).not.toBeInTheDocument()
  })

  it('shows error on API failure', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            id: 1,
            name: 'Admin',
            email: 'admin@example.com',
            phone: '555-0001',
            role: 'admin',
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        new Response('internal error', { status: 500 }),
      )

    render(
      <MemoryRouter initialEntries={['/members']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="/members" element={<Members />} />
            </Route>
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to load members')
    })
  })

  it('is not visible to regular users (nav link hidden)', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({
          id: 1,
          name: 'Regular User',
          email: 'user@example.com',
          phone: '555-0001',
          role: 'user',
        }),
        { status: 200 },
      ),
    )

    render(
      <MemoryRouter initialEntries={['/']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route index element={<div>Home Content</div>} />
              <Route path="members" element={<Members />} />
            </Route>
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByText('Home Content')).toBeInTheDocument()
    })

    // Members link should not be in the nav
    const nav = screen.getByRole('navigation')
    expect(within(nav).queryByText('Members')).not.toBeInTheDocument()
  })
})
