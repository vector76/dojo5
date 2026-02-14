import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { ProtectedRoute } from './ProtectedRoute'
import { setToken } from '../api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

function renderProtected(initialPath = '/') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<div>Login Page</div>} />
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <div>Protected Content</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('ProtectedRoute', () => {
  it('redirects to /login when not authenticated', async () => {
    renderProtected()

    await waitFor(() => {
      expect(screen.getByText('Login Page')).toBeInTheDocument()
    })
    expect(screen.queryByText('Protected Content')).not.toBeInTheDocument()
  })

  it('shows loading state while checking auth', () => {
    setToken('some-token')
    // fetch will hang (never resolves), so loading stays true
    vi.spyOn(globalThis, 'fetch').mockReturnValue(new Promise(() => {}))

    renderProtected()

    expect(screen.getByText('Loading...')).toBeInTheDocument()
    expect(screen.queryByText('Protected Content')).not.toBeInTheDocument()
    expect(screen.queryByText('Login Page')).not.toBeInTheDocument()
  })

  it('renders children when authenticated', async () => {
    setToken('valid-token')
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

    renderProtected()

    await waitFor(() => {
      expect(screen.getByText('Protected Content')).toBeInTheDocument()
    })
    expect(screen.queryByText('Login Page')).not.toBeInTheDocument()
  })

  it('redirects to /login when token is expired (401)', async () => {
    setToken('expired-token')
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response('unauthorized', { status: 401 }),
    )

    renderProtected()

    await waitFor(() => {
      expect(screen.getByText('Login Page')).toBeInTheDocument()
    })
    expect(screen.queryByText('Protected Content')).not.toBeInTheDocument()
  })
})
