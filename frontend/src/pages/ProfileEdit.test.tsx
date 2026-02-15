import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { Layout } from '../components/Layout'
import { ProfileEdit } from './ProfileEdit'
import { setToken } from '../api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

const profileData = {
  id: 5,
  name: 'Regular User',
  email: 'user@example.com',
  phone: '555-5555',
  role: 'user',
  emergency_contact: 'Mom 555-0000',
}

function renderProfileEdit(role = 'user') {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch')
    // GET /api/auth/me
    .mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 5, name: 'Regular User', email: 'user@example.com', phone: '555-5555', role }), { status: 200 }),
    )
    // GET /api/users/5
    .mockResolvedValueOnce(
      new Response(JSON.stringify(profileData), { status: 200 }),
    )
  return render(
    <MemoryRouter initialEntries={['/profile']}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route path="profile" element={<ProfileEdit />} />
          </Route>
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('ProfileEdit', () => {
  it('renders profile edit form with current values', async () => {
    renderProfileEdit()

    await waitFor(() => {
      expect(screen.getByLabelText('Name')).toHaveValue('Regular User')
    })
    expect(screen.getByLabelText('Email')).toHaveValue('user@example.com')
    expect(screen.getByLabelText('Phone')).toHaveValue('555-5555')
    expect(screen.getByLabelText('Emergency Contact')).toHaveValue('Mom 555-0000')
  })

  it('does not show admin-only fields (role, membership, etc.)', async () => {
    renderProfileEdit()

    await waitFor(() => {
      expect(screen.getByLabelText('Name')).toBeInTheDocument()
    })
    expect(screen.queryByLabelText('Role')).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Membership Type')).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Membership Status')).not.toBeInTheDocument()
  })

  it('saves profile changes', async () => {
    const user = userEvent.setup()
    renderProfileEdit()

    await waitFor(() => {
      expect(screen.getByLabelText('Name')).toHaveValue('Regular User')
    })

    await user.clear(screen.getByLabelText('Phone'))
    await user.type(screen.getByLabelText('Phone'), '555-9999')

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ ...profileData, phone: '555-9999' }), { status: 200 }),
    )

    await user.click(screen.getByText('Save'))

    await waitFor(() => {
      expect(screen.getByText('Profile updated successfully')).toBeInTheDocument()
    })
  })

  it('shows error on save failure', async () => {
    const user = userEvent.setup()
    renderProfileEdit()

    await waitFor(() => {
      expect(screen.getByLabelText('Name')).toHaveValue('Regular User')
    })

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response('server error', { status: 500 }),
    )

    await user.click(screen.getByText('Save'))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to save profile')
    })
  })

  it('shows error when profile fails to load', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ id: 5, name: 'User', email: 'u@example.com', phone: '555', role: 'user' }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response('error', { status: 500 }),
      )

    render(
      <MemoryRouter initialEntries={['/profile']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="profile" element={<ProfileEdit />} />
            </Route>
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to load profile')
    })
  })
})
