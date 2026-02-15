import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { Layout } from '../components/Layout'
import { AddMember } from './AddMember'
import { setToken } from '../api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

function renderAddMember() {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
    new Response(JSON.stringify({ id: 1, name: 'Admin', email: 'admin@example.com', phone: '555-0001', role: 'admin' }), { status: 200 }),
  )
  return render(
    <MemoryRouter initialEntries={['/members/new']}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route path="members/new" element={<AddMember />} />
            <Route path="members" element={<div>Members List</div>} />
          </Route>
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('AddMember', () => {
  it('renders the add member form', async () => {
    renderAddMember()

    await waitFor(() => {
      expect(screen.getByText('Add Member')).toBeInTheDocument()
    })
    expect(screen.getByLabelText('Name')).toBeInTheDocument()
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
    expect(screen.getByLabelText('Phone')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
    expect(screen.getByLabelText('Role')).toBeInTheDocument()
  })

  it('creates a member and navigates to members list', async () => {
    const user = userEvent.setup()
    renderAddMember()

    await waitFor(() => {
      expect(screen.getByText('Add Member')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('Name'), 'New Person')
    await user.type(screen.getByLabelText('Email'), 'new@example.com')
    await user.type(screen.getByLabelText('Phone'), '555-1234')
    await user.type(screen.getByLabelText('Password'), 'secret123')
    await user.selectOptions(screen.getByLabelText('Role'), 'instructor')

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 5, name: 'New Person', email: 'new@example.com', phone: '555-1234', role: 'instructor' }), { status: 201 }),
    )

    await user.click(screen.getByText('Create Member'))

    await waitFor(() => {
      expect(screen.getByText('Members List')).toBeInTheDocument()
    })
  })

  it('shows error on API failure', async () => {
    const user = userEvent.setup()
    renderAddMember()

    await waitFor(() => {
      expect(screen.getByText('Add Member')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('Name'), 'New Person')
    await user.type(screen.getByLabelText('Email'), 'new@example.com')
    await user.type(screen.getByLabelText('Phone'), '555-1234')
    await user.type(screen.getByLabelText('Password'), 'secret123')

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response('email already exists', { status: 409 }),
    )

    await user.click(screen.getByText('Create Member'))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to create member')
    })
  })

  it('defaults role to user', async () => {
    renderAddMember()

    await waitFor(() => {
      expect(screen.getByLabelText('Role')).toHaveValue('user')
    })
  })
})
