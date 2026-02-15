import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { Layout } from '../components/Layout'
import { Attendance } from './Attendance'
import { setToken } from '../api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

const admin = { id: 1, name: 'Admin User', email: 'admin@example.com', phone: '555-0001', role: 'admin' }

const classTypes = [
  { id: 1, name: 'Yoga', description: 'Yoga class' },
  { id: 2, name: 'Karate', description: 'Karate class' },
]

const members = [
  { id: 1, name: 'Admin User', email: 'admin@example.com', phone: '555-0001', role: 'admin' },
  { id: 2, name: 'Instructor One', email: 'inst@example.com', phone: '555-0002', role: 'instructor' },
  { id: 3, name: 'Alice Student', email: 'alice@example.com', phone: '555-0003', role: 'user' },
  { id: 4, name: 'Bob Student', email: 'bob@example.com', phone: '555-0004', role: 'user' },
]

const classes = [
  { id: 10, class_type_id: 1, instructor_id: 2, start_time: '2026-03-01T10:00:00Z', duration_minutes: 60, capacity: 20 },
  { id: 11, class_type_id: 2, instructor_id: 1, start_time: '2026-03-02T14:00:00Z', duration_minutes: 45, capacity: 15 },
]

const attendanceRecords = [
  { id: 100, class_id: 10, user_id: 3, checked_in_at: '2026-03-01T10:05:00Z' },
]

function renderPage() {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch')
    // GET /api/auth/me
    .mockResolvedValueOnce(new Response(JSON.stringify(admin), { status: 200 }))
    // GET /api/classes
    .mockResolvedValueOnce(new Response(JSON.stringify(classes), { status: 200 }))
    // GET /api/class-types
    .mockResolvedValueOnce(new Response(JSON.stringify(classTypes), { status: 200 }))
    // GET /api/users
    .mockResolvedValueOnce(new Response(JSON.stringify(members), { status: 200 }))
  return render(
    <MemoryRouter initialEntries={['/attendance']}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route path="attendance" element={<Attendance />} />
          </Route>
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('Attendance', () => {
  it('displays the page with class selector', async () => {
    renderPage()

    await waitFor(() => {
      expect(screen.getByLabelText('Select Class')).toBeInTheDocument()
    })

    expect(screen.getByRole('heading', { name: 'Attendance' })).toBeInTheDocument()
  })

  it('shows class options in the selector', async () => {
    renderPage()

    await waitFor(() => {
      expect(screen.getByLabelText('Select Class')).toBeInTheDocument()
    })

    const select = screen.getByLabelText('Select Class')
    const options = within(select).getAllByRole('option')
    // placeholder + 2 classes
    expect(options).toHaveLength(3)
  })

  it('loads attendance when a class is selected', async () => {
    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByLabelText('Select Class')).toBeInTheDocument()
    })

    // Mock GET /api/classes/10/attendance
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify(attendanceRecords), { status: 200 }),
    )

    await user.selectOptions(screen.getByLabelText('Select Class'), '10')

    await waitFor(() => {
      expect(screen.getByText('Alice Student')).toBeInTheDocument()
    })

    const checkedInSection = screen.getByRole('region', { name: 'Checked In' })
    expect(within(checkedInSection).getByText('Alice Student')).toBeInTheDocument()
  })

  it('shows empty attendance state', async () => {
    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByLabelText('Select Class')).toBeInTheDocument()
    })

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify([]), { status: 200 }),
    )

    await user.selectOptions(screen.getByLabelText('Select Class'), '10')

    await waitFor(() => {
      expect(screen.getByText('No one checked in yet.')).toBeInTheDocument()
    })
  })

  it('checks in a member', async () => {
    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByLabelText('Select Class')).toBeInTheDocument()
    })

    // Load attendance for class 10 (Alice already checked in)
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify(attendanceRecords), { status: 200 }),
    )

    await user.selectOptions(screen.getByLabelText('Select Class'), '10')

    await waitFor(() => {
      expect(screen.getByText('Alice Student')).toBeInTheDocument()
    })

    // Alice is already checked in, so she shouldn't be in the member select
    const checkInSection = screen.getByRole('region', { name: 'Check In' })
    const memberSelect = within(checkInSection).getByLabelText('Member')
    expect(within(memberSelect).queryByText('Alice Student')).not.toBeInTheDocument()

    // Select Bob and check in
    await user.selectOptions(memberSelect, '4')

    const newRecord = { id: 101, class_id: 10, user_id: 4, checked_in_at: '2026-03-01T10:10:00Z' }
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify(newRecord), { status: 201 }),
    )

    await user.click(within(checkInSection).getByRole('button', { name: 'Check In' }))

    await waitFor(() => {
      expect(screen.getByText('Bob Student')).toBeInTheDocument()
    })
  })

  it('shows error when check-in fails', async () => {
    const user = userEvent.setup()
    renderPage()

    await waitFor(() => {
      expect(screen.getByLabelText('Select Class')).toBeInTheDocument()
    })

    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
      new Response(JSON.stringify([]), { status: 200 }),
    )

    await user.selectOptions(screen.getByLabelText('Select Class'), '10')

    await waitFor(() => {
      expect(screen.getByText('No one checked in yet.')).toBeInTheDocument()
    })

    const checkInSection = screen.getByRole('region', { name: 'Check In' })
    const memberSelect = within(checkInSection).getByLabelText('Member')
    await user.selectOptions(memberSelect, '3')

    vi.spyOn(globalThis, 'fetch').mockRejectedValueOnce(new Error('Server error'))

    await user.click(within(checkInSection).getByRole('button', { name: 'Check In' }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to check in member')
    })
  })

  it('shows error when loading data fails', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(JSON.stringify(admin), { status: 200 }))
      .mockRejectedValueOnce(new Error('Network error'))
    render(
      <MemoryRouter initialEntries={['/attendance']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="attendance" element={<Attendance />} />
            </Route>
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to load data')
    })
  })
})
