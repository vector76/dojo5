import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { Setup } from './Setup'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

function renderSetup() {
  return render(
    <MemoryRouter initialEntries={['/setup']}>
      <AuthProvider>
        <Routes>
          <Route path="/setup" element={<Setup />} />
          <Route path="/" element={<div>Dashboard</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('Setup', () => {
  it('renders the setup form', () => {
    renderSetup()

    expect(screen.getByText('Welcome to Dojo CRM')).toBeInTheDocument()
    expect(screen.getByLabelText('Name')).toBeInTheDocument()
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
    expect(screen.getByLabelText('Phone')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
    expect(screen.getByText('Create Admin Account')).toBeInTheDocument()
  })

  it('submits setup form and stores token', async () => {
    const user = userEvent.setup()

    // Mock window.location.href assignment
    const original = window.location.href
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { ...window.location, href: original },
    })

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ token: 'admin-jwt' }), { status: 201 }),
    )

    renderSetup()

    await user.type(screen.getByLabelText('Name'), 'Admin User')
    await user.type(screen.getByLabelText('Email'), 'admin@example.com')
    await user.type(screen.getByLabelText('Phone'), '555-0001')
    await user.type(screen.getByLabelText('Password'), 'securepass')
    await user.click(screen.getByText('Create Admin Account'))

    await waitFor(() => {
      expect(localStorage.getItem('auth_token')).toBe('admin-jwt')
    })

    // Verify the fetch was called with correct data
    const [url, init] = vi.mocked(fetch).mock.calls[0]
    expect(url).toBe('/api/auth/setup')
    expect(init?.method).toBe('POST')
    const body = JSON.parse(init?.body as string)
    expect(body).toEqual({
      name: 'Admin User',
      email: 'admin@example.com',
      phone: '555-0001',
      password: 'securepass',
    })

    // No cleanup needed — jsdom resets between tests
  })

  it('shows error when setup already completed (409)', async () => {
    const user = userEvent.setup()

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response('setup already completed', { status: 409 }),
    )

    renderSetup()

    await user.type(screen.getByLabelText('Name'), 'Admin')
    await user.type(screen.getByLabelText('Email'), 'admin@example.com')
    await user.type(screen.getByLabelText('Phone'), '555-0001')
    await user.type(screen.getByLabelText('Password'), 'pass')
    await user.click(screen.getByText('Create Admin Account'))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Setup has already been completed')
    })
  })

  it('shows error on server error', async () => {
    const user = userEvent.setup()

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response('name, email, and phone are required', { status: 400 }),
    )

    renderSetup()

    await user.type(screen.getByLabelText('Name'), 'Admin')
    await user.type(screen.getByLabelText('Email'), 'admin@example.com')
    await user.type(screen.getByLabelText('Phone'), '555-0001')
    await user.type(screen.getByLabelText('Password'), 'pass')
    await user.click(screen.getByText('Create Admin Account'))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
  })
})
