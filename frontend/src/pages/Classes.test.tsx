import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { Layout } from '../components/Layout'
import { Classes } from './Classes'
import { setToken } from '../api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

const classTypes = [
  { id: 1, name: 'Yoga', description: 'Yoga class' },
  { id: 2, name: 'Karate', description: 'Karate class' },
]

const members = [
  { id: 1, name: 'Admin User', email: 'admin@example.com', phone: '555-0001', role: 'admin' },
  { id: 2, name: 'Instructor One', email: 'inst@example.com', phone: '555-0002', role: 'instructor' },
  { id: 3, name: 'Regular User', email: 'user@example.com', phone: '555-0003', role: 'user' },
]

const classes = [
  { id: 1, class_type_id: 1, instructor_id: 2, start_time: '2026-03-01T10:00:00Z', duration_minutes: 60, capacity: 20 },
  { id: 2, class_type_id: 2, instructor_id: 1, start_time: '2026-03-02T14:00:00Z', duration_minutes: 45, capacity: 15 },
]

function renderAsAdmin() {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch')
    // GET /api/auth/me
    .mockResolvedValueOnce(
      new Response(JSON.stringify(members[0]), { status: 200 }),
    )
    // GET /api/classes
    .mockResolvedValueOnce(
      new Response(JSON.stringify(classes), { status: 200 }),
    )
    // GET /api/class-types
    .mockResolvedValueOnce(
      new Response(JSON.stringify(classTypes), { status: 200 }),
    )
    // GET /api/users (for instructor list)
    .mockResolvedValueOnce(
      new Response(JSON.stringify(members), { status: 200 }),
    )
  return render(
    <MemoryRouter initialEntries={['/classes']}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route path="classes" element={<Classes />} />
          </Route>
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

function renderAsUser() {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch')
    // GET /api/auth/me
    .mockResolvedValueOnce(
      new Response(JSON.stringify(members[2]), { status: 200 }),
    )
    // GET /api/classes
    .mockResolvedValueOnce(
      new Response(JSON.stringify(classes), { status: 200 }),
    )
    // GET /api/class-types
    .mockResolvedValueOnce(
      new Response(JSON.stringify(classTypes), { status: 200 }),
    )
  return render(
    <MemoryRouter initialEntries={['/classes']}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route path="classes" element={<Classes />} />
          </Route>
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('Classes', () => {
  it('displays class schedule for all users', async () => {
    renderAsUser()

    await waitFor(() => {
      expect(screen.getByText('Class Schedule')).toBeInTheDocument()
    })

    expect(screen.getByText('Yoga')).toBeInTheDocument()
    expect(screen.getByText('Karate')).toBeInTheDocument()
    expect(screen.getByText('60 min')).toBeInTheDocument()
    expect(screen.getByText('45 min')).toBeInTheDocument()
    expect(screen.getByText('20')).toBeInTheDocument()
    expect(screen.getByText('15')).toBeInTheDocument()
  })

  it('does not show admin actions for regular users', async () => {
    renderAsUser()

    await waitFor(() => {
      expect(screen.getByText('Class Schedule')).toBeInTheDocument()
    })

    expect(screen.queryByText('Add Class')).not.toBeInTheDocument()
    expect(screen.queryByText('Edit')).not.toBeInTheDocument()
    expect(screen.queryByText('Delete')).not.toBeInTheDocument()
  })

  it('shows admin actions for admin users', async () => {
    renderAsAdmin()

    await waitFor(() => {
      expect(screen.getByText('Class Schedule')).toBeInTheDocument()
    })

    expect(screen.getByText('Add Class')).toBeInTheDocument()
    const editButtons = screen.getAllByText('Edit')
    expect(editButtons).toHaveLength(2)
    const deleteButtons = screen.getAllByText('Delete')
    expect(deleteButtons).toHaveLength(2)
  })

  it('shows instructor names for admin users', async () => {
    renderAsAdmin()

    await waitFor(() => {
      expect(screen.getByText('Instructor One')).toBeInTheDocument()
    })

    // "Admin User" appears in sidebar and table — check via table rows
    const table = screen.getByRole('table')
    expect(within(table).getByText('Admin User')).toBeInTheDocument()
  })

  it('shows date range filter form', async () => {
    renderAsUser()

    await waitFor(() => {
      expect(screen.getByText('Class Schedule')).toBeInTheDocument()
    })

    expect(screen.getByLabelText('From')).toBeInTheDocument()
    expect(screen.getByLabelText('To')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Filter' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Clear' })).toBeInTheDocument()
  })

  it('filters classes by date range', async () => {
    const user = userEvent.setup()
    renderAsUser()

    await waitFor(() => {
      expect(screen.getByText('Class Schedule')).toBeInTheDocument()
    })

    const filteredClasses = [classes[0]]
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify(filteredClasses), { status: 200 }),
    )

    await user.type(screen.getByLabelText('From'), '2026-03-01')
    await user.type(screen.getByLabelText('To'), '2026-03-01')
    await user.click(screen.getByRole('button', { name: 'Filter' }))

    await waitFor(() => {
      expect(screen.getByText('Yoga')).toBeInTheDocument()
    })
    expect(screen.queryByText('Karate')).not.toBeInTheDocument()
  })

  it('clears filter and reloads all classes', async () => {
    const user = userEvent.setup()
    renderAsUser()

    await waitFor(() => {
      expect(screen.getByText('Class Schedule')).toBeInTheDocument()
    })

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify(classes), { status: 200 }),
    )

    await user.click(screen.getByRole('button', { name: 'Clear' }))

    await waitFor(() => {
      expect(screen.getByText('Yoga')).toBeInTheDocument()
      expect(screen.getByText('Karate')).toBeInTheDocument()
    })
  })

  it('shows create class form when admin clicks Add Class', async () => {
    const user = userEvent.setup()
    renderAsAdmin()

    await waitFor(() => {
      expect(screen.getByText('Add Class')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Add Class'))

    expect(screen.getByLabelText('Create class')).toBeInTheDocument()
    expect(screen.getByLabelText('Class Type')).toBeInTheDocument()
    expect(screen.getByLabelText('Instructor')).toBeInTheDocument()
    expect(screen.getByLabelText('Start Time')).toBeInTheDocument()
    expect(screen.getByLabelText('Duration (minutes)')).toBeInTheDocument()
    expect(screen.getByLabelText('Capacity')).toBeInTheDocument()
  })

  it('creates a new class', async () => {
    const user = userEvent.setup()
    renderAsAdmin()

    await waitFor(() => {
      expect(screen.getByText('Add Class')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Add Class'))

    const form = screen.getByLabelText('Create class')
    const startTimeInput = within(form).getByLabelText('Start Time')
    await user.clear(within(form).getByLabelText('Duration (minutes)'))
    await user.type(within(form).getByLabelText('Duration (minutes)'), '90')
    await user.clear(within(form).getByLabelText('Capacity'))
    await user.type(within(form).getByLabelText('Capacity'), '25')
    await user.type(startTimeInput, '2026-04-01T09:00')

    const newClass = { id: 3, class_type_id: 1, instructor_id: 2, start_time: '2026-04-01T09:00:00Z', duration_minutes: 90, capacity: 25 }
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify(newClass), { status: 201 }),
    )

    await user.click(within(form).getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(screen.queryByLabelText('Create class')).not.toBeInTheDocument()
    })
    expect(screen.getByText('90 min')).toBeInTheDocument()
    expect(screen.getByText('25')).toBeInTheDocument()
  })

  it('shows edit class form when admin clicks Edit', async () => {
    const user = userEvent.setup()
    renderAsAdmin()

    await waitFor(() => {
      expect(screen.getByText('Class Schedule')).toBeInTheDocument()
    })

    const editButtons = screen.getAllByText('Edit')
    await user.click(editButtons[0])

    expect(screen.getByLabelText('Edit class')).toBeInTheDocument()
  })

  it('shows error when class creation fails', async () => {
    const user = userEvent.setup()
    renderAsAdmin()

    await waitFor(() => {
      expect(screen.getByText('Add Class')).toBeInTheDocument()
    })

    await user.click(screen.getByText('Add Class'))

    const form = screen.getByLabelText('Create class')
    await user.type(within(form).getByLabelText('Start Time'), '2026-04-01T09:00')

    vi.spyOn(globalThis, 'fetch').mockRejectedValueOnce(new Error('Server error'))

    await user.click(within(form).getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to create class')
    })
  })

  it('deletes a class when admin confirms', async () => {
    const user = userEvent.setup()
    vi.spyOn(globalThis, 'confirm').mockReturnValue(true)
    renderAsAdmin()

    await waitFor(() => {
      expect(screen.getByText('Class Schedule')).toBeInTheDocument()
    })

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(null, { status: 204 }),
    )

    const deleteButtons = screen.getAllByText('Delete')
    await user.click(deleteButtons[0])

    await waitFor(() => {
      const rows = screen.getAllByRole('row')
      // Header + 1 remaining class
      expect(rows).toHaveLength(2)
    })
  })

  it('shows empty state when no classes exist', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify(members[2]), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify([]), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(classTypes), { status: 200 }),
      )
    render(
      <MemoryRouter initialEntries={['/classes']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="classes" element={<Classes />} />
            </Route>
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByText('No classes found')).toBeInTheDocument()
    })
  })

  it('shows error when loading fails', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(
        new Response(JSON.stringify(members[2]), { status: 200 }),
      )
      .mockRejectedValueOnce(new Error('Network error'))
    render(
      <MemoryRouter initialEntries={['/classes']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="classes" element={<Classes />} />
            </Route>
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to load class schedule')
    })
  })
})
