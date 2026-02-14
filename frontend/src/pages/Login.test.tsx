import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { Login } from './Login'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

function renderLogin() {
  return render(
    <MemoryRouter initialEntries={['/login']}>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/setup" element={<div>Setup Page</div>} />
          <Route path="/" element={<div>Dashboard</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

function mockSetupStatus(needsSetup: boolean) {
  return vi.spyOn(globalThis, 'fetch').mockResolvedValue(
    new Response(JSON.stringify({ needs_setup: needsSetup }), { status: 200 }),
  )
}

describe('Login', () => {
  it('shows login form when setup is complete', async () => {
    mockSetupStatus(false)

    renderLogin()

    await waitFor(() => {
      expect(screen.getByText('Login')).toBeInTheDocument()
    })
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
    expect(screen.getByText('Log in')).toBeInTheDocument()
  })

  it('redirects to /setup when setup is needed', async () => {
    mockSetupStatus(true)

    renderLogin()

    await waitFor(() => {
      expect(screen.getByText('Setup Page')).toBeInTheDocument()
    })
    expect(screen.queryByText('Login')).not.toBeInTheDocument()
  })

  it('shows loading state while checking setup status', () => {
    vi.spyOn(globalThis, 'fetch').mockReturnValue(new Promise(() => {}))

    renderLogin()

    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('submits login and redirects to dashboard on success', async () => {
    const user = userEvent.setup()
    vi.spyOn(globalThis, 'fetch')
      // setup-status check
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ needs_setup: false }), { status: 200 }),
      )
      // POST /api/auth/login
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ token: 'new-jwt' }), { status: 200 }),
      )
      // GET /api/auth/me
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

    renderLogin()

    await waitFor(() => {
      expect(screen.getByText('Login')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('Email'), 'admin@example.com')
    await user.type(screen.getByLabelText('Password'), 'password')
    await user.click(screen.getByText('Log in'))

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument()
    })
    expect(localStorage.getItem('auth_token')).toBe('new-jwt')
  })

  it('shows error on invalid credentials', async () => {
    const user = userEvent.setup()
    vi.spyOn(globalThis, 'fetch')
      // setup-status check
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ needs_setup: false }), { status: 200 }),
      )
      // POST /api/auth/login - 401
      .mockResolvedValueOnce(
        new Response('invalid credentials', { status: 401 }),
      )

    renderLogin()

    await waitFor(() => {
      expect(screen.getByText('Login')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('Email'), 'bad@example.com')
    await user.type(screen.getByLabelText('Password'), 'wrong')
    await user.click(screen.getByText('Log in'))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Invalid email or password')
    })
  })
})
