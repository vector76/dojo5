import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { Layout } from '../components/Layout'
import { ClassTypes } from './ClassTypes'
import { setToken } from '../api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

const mockClassTypes = [
  { id: 1, name: 'Hatha Yoga', description: 'Traditional yoga practice' },
  { id: 2, name: 'Vinyasa', description: 'Flow-based yoga' },
  { id: 3, name: 'Meditation', description: null },
]

function renderClassTypes() {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch')
    // GET /api/auth/me
    .mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 1, name: 'Admin', email: 'admin@example.com', phone: '555-0001', role: 'admin' }), { status: 200 }),
    )
    // GET /api/class-types
    .mockResolvedValueOnce(
      new Response(JSON.stringify(mockClassTypes), { status: 200 }),
    )
  return render(
    <MemoryRouter initialEntries={['/class-types']}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route path="class-types" element={<ClassTypes />} />
          </Route>
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('ClassTypes', () => {
  it('renders class type list', async () => {
    renderClassTypes()

    await waitFor(() => {
      expect(screen.getByText('Hatha Yoga')).toBeInTheDocument()
    })
    expect(screen.getByText('Vinyasa')).toBeInTheDocument()
    expect(screen.getByText('Meditation')).toBeInTheDocument()
    expect(screen.getByText('Traditional yoga practice')).toBeInTheDocument()
    expect(screen.getByText('Flow-based yoga')).toBeInTheDocument()
  })

  it('shows empty state when no class types', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ id: 1, name: 'Admin', email: 'admin@example.com', phone: '555-0001', role: 'admin' }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify([]), { status: 200 }),
      )

    render(
      <MemoryRouter initialEntries={['/class-types']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="class-types" element={<ClassTypes />} />
            </Route>
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByText('No class types found')).toBeInTheDocument()
    })
  })

  it('creates a new class type', async () => {
    const user = userEvent.setup()
    renderClassTypes()

    await waitFor(() => {
      expect(screen.getByText('Hatha Yoga')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Add Class Type'))

    expect(screen.getByRole('form', { name: 'Create class type' })).toBeInTheDocument()

    await user.type(screen.getByLabelText('Name'), 'Yin Yoga')
    await user.type(screen.getByLabelText('Description'), 'Slow-paced yoga')

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 4, name: 'Yin Yoga', description: 'Slow-paced yoga' }), { status: 201 }),
    )

    await user.click(screen.getByText('Save'))

    await waitFor(() => {
      expect(screen.getByText('Yin Yoga')).toBeInTheDocument()
    })
    // Form should be closed
    expect(screen.queryByRole('form')).not.toBeInTheDocument()
  })

  it('edits an existing class type', async () => {
    const user = userEvent.setup()
    renderClassTypes()

    await waitFor(() => {
      expect(screen.getByText('Hatha Yoga')).toBeInTheDocument()
    })

    // Click edit on the first class type
    const editButtons = screen.getAllByText('Edit')
    await user.click(editButtons[0])

    expect(screen.getByRole('form', { name: 'Edit class type' })).toBeInTheDocument()
    expect(screen.getByLabelText('Name')).toHaveValue('Hatha Yoga')
    expect(screen.getByLabelText('Description')).toHaveValue('Traditional yoga practice')

    await user.clear(screen.getByLabelText('Name'))
    await user.type(screen.getByLabelText('Name'), 'Hatha Flow')

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 1, name: 'Hatha Flow', description: 'Traditional yoga practice' }), { status: 200 }),
    )

    await user.click(screen.getByText('Save'))

    await waitFor(() => {
      expect(screen.getByText('Hatha Flow')).toBeInTheDocument()
    })
    expect(screen.queryByText('Hatha Yoga')).not.toBeInTheDocument()
  })

  it('deletes a class type', async () => {
    const user = userEvent.setup()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    renderClassTypes()

    await waitFor(() => {
      expect(screen.getByText('Hatha Yoga')).toBeInTheDocument()
    })

    const deleteButtons = screen.getAllByText('Delete')

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(null, { status: 204 }),
    )

    await user.click(deleteButtons[0])

    await waitFor(() => {
      expect(screen.queryByText('Hatha Yoga')).not.toBeInTheDocument()
    })
    // Other items should still be there
    expect(screen.getByText('Vinyasa')).toBeInTheDocument()
  })

  it('cancels form without saving', async () => {
    const user = userEvent.setup()
    renderClassTypes()

    await waitFor(() => {
      expect(screen.getByText('Hatha Yoga')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Add Class Type'))
    expect(screen.getByRole('form')).toBeInTheDocument()

    await user.type(screen.getByLabelText('Name'), 'Test')

    await user.click(screen.getByText('Cancel'))

    expect(screen.queryByRole('form')).not.toBeInTheDocument()
  })

  it('shows error on create failure', async () => {
    const user = userEvent.setup()
    renderClassTypes()

    await waitFor(() => {
      expect(screen.getByText('Hatha Yoga')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Add Class Type'))
    await user.type(screen.getByLabelText('Name'), 'Bad Type')

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response('server error', { status: 500 }),
    )

    await user.click(screen.getByText('Save'))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to create class type')
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
      <MemoryRouter initialEntries={['/class-types']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="class-types" element={<ClassTypes />} />
            </Route>
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to load class types')
    })
  })

  it('shows error on edit failure', async () => {
    const user = userEvent.setup()
    renderClassTypes()

    await waitFor(() => {
      expect(screen.getByText('Hatha Yoga')).toBeInTheDocument()
    })

    const editButtons = screen.getAllByText('Edit')
    await user.click(editButtons[0])

    await user.clear(screen.getByLabelText('Name'))
    await user.type(screen.getByLabelText('Name'), 'Updated')

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response('server error', { status: 500 }),
    )

    await user.click(screen.getByText('Save'))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to update class type')
    })
  })
})
