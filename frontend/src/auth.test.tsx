import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AuthProvider, useAuth } from './auth'
import { setToken } from './api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

function TestConsumer() {
  const { user, loading, login, logout } = useAuth()

  if (loading) return <div>loading</div>

  return (
    <div>
      <div data-testid="user">{user ? user.name : 'none'}</div>
      <div data-testid="role">{user ? user.role : ''}</div>
      <button onClick={() => login('test@example.com', 'pass')}>login</button>
      <button onClick={logout}>logout</button>
    </div>
  )
}

function renderWithAuth() {
  return render(
    <AuthProvider>
      <TestConsumer />
    </AuthProvider>,
  )
}

describe('AuthProvider', () => {
  it('starts with no user when no token is stored', async () => {
    vi.spyOn(globalThis, 'fetch')

    renderWithAuth()

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('none')
    })

    // Should not call /api/auth/me when there is no token
    expect(fetch).not.toHaveBeenCalled()
  })

  it('fetches current user on mount when token exists', async () => {
    setToken('existing-token')

    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
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

    renderWithAuth()

    expect(screen.getByText('loading')).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('Admin')
    })
    expect(screen.getByTestId('role')).toHaveTextContent('admin')
  })

  it('clears token and sets user to null on 401 during initial fetch', async () => {
    setToken('expired-token')

    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response('unauthorized', { status: 401 }),
    )

    renderWithAuth()

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('none')
    })

    expect(localStorage.getItem('auth_token')).toBeNull()
  })

  it('login stores token and fetches user', async () => {
    const user = userEvent.setup()

    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        // POST /api/auth/login
        new Response(JSON.stringify({ token: 'new-jwt' }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        // GET /api/auth/me
        new Response(
          JSON.stringify({
            id: 2,
            name: 'Jane',
            email: 'jane@example.com',
            phone: '555-0002',
            role: 'instructor',
          }),
          { status: 200 },
        ),
      )

    renderWithAuth()

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('none')
    })

    await user.click(screen.getByText('login'))

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('Jane')
    })
    expect(screen.getByTestId('role')).toHaveTextContent('instructor')
    expect(localStorage.getItem('auth_token')).toBe('new-jwt')
  })

  it('logout clears token and user', async () => {
    const user = userEvent.setup()
    setToken('existing-token')

    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
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

    renderWithAuth()

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('Admin')
    })

    await user.click(screen.getByText('logout'))

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('none')
    })
    expect(localStorage.getItem('auth_token')).toBeNull()
  })

  it('throws when useAuth is used outside AuthProvider', () => {
    // Suppress expected error output
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})

    function BadComponent() {
      useAuth()
      return null
    }

    expect(() => render(<BadComponent />)).toThrow(
      'useAuth must be used within an AuthProvider',
    )

    spy.mockRestore()
  })
})
