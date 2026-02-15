import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../auth'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { Layout } from '../components/Layout'
import { MyAttendance } from './MyAttendance'
import { setToken } from '../api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

const regularUser = { id: 3, name: 'Alice Student', email: 'alice@example.com', phone: '555-0003', role: 'user' }

const classTypes = [
  { id: 1, name: 'Yoga', description: 'Yoga class' },
  { id: 2, name: 'Karate', description: 'Karate class' },
]

const classes = [
  { id: 10, class_type_id: 1, instructor_id: 2, start_time: '2026-03-01T10:00:00Z', duration_minutes: 60, capacity: 20 },
  { id: 11, class_type_id: 2, instructor_id: 1, start_time: '2026-03-02T14:00:00Z', duration_minutes: 45, capacity: 15 },
]

const attendanceRecords = [
  { id: 100, class_id: 10, user_id: 3, checked_in_at: '2026-03-01T10:05:00Z' },
  { id: 101, class_id: 11, user_id: 3, checked_in_at: '2026-03-02T14:02:00Z' },
]

function renderPage() {
  setToken('valid-token')
  vi.spyOn(globalThis, 'fetch')
    // GET /api/auth/me
    .mockResolvedValueOnce(new Response(JSON.stringify(regularUser), { status: 200 }))
    // GET /api/users/3/attendance
    .mockResolvedValueOnce(new Response(JSON.stringify(attendanceRecords), { status: 200 }))
    // GET /api/classes
    .mockResolvedValueOnce(new Response(JSON.stringify(classes), { status: 200 }))
    // GET /api/class-types
    .mockResolvedValueOnce(new Response(JSON.stringify(classTypes), { status: 200 }))
  return render(
    <MemoryRouter initialEntries={['/my-attendance']}>
      <AuthProvider>
        <Routes>
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route path="my-attendance" element={<MyAttendance />} />
          </Route>
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('MyAttendance', () => {
  it('displays attendance history with class info', async () => {
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Total classes attended: 2')).toBeInTheDocument()
    })

    expect(screen.getByRole('heading', { name: 'My Attendance' })).toBeInTheDocument()
    expect(screen.getByText('Yoga')).toBeInTheDocument()
    expect(screen.getByText('Karate')).toBeInTheDocument()
  })

  it('shows empty state when no attendance records', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(JSON.stringify(regularUser), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify([]), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify(classes), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify(classTypes), { status: 200 }))
    render(
      <MemoryRouter initialEntries={['/my-attendance']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="my-attendance" element={<MyAttendance />} />
            </Route>
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByText('No attendance records yet.')).toBeInTheDocument()
    })
    expect(screen.getByText('Total classes attended: 0')).toBeInTheDocument()
  })

  it('shows total classes attended count', async () => {
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Total classes attended: 2')).toBeInTheDocument()
    })
  })

  it('shows error when loading fails', async () => {
    setToken('valid-token')
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(JSON.stringify(regularUser), { status: 200 }))
      .mockRejectedValueOnce(new Error('Network error'))
    render(
      <MemoryRouter initialEntries={['/my-attendance']}>
        <AuthProvider>
          <Routes>
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="my-attendance" element={<MyAttendance />} />
            </Route>
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to load attendance history')
    })
  })
})
