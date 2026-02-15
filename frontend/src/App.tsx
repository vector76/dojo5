import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from './auth'
import { ErrorBoundary } from './components/ErrorBoundary'
import { ToastProvider } from './components/Toast'
import { ProtectedRoute } from './components/ProtectedRoute'
import { Layout } from './components/Layout'
import { Dashboard } from './pages/Dashboard'
import { Login } from './pages/Login'
import { Setup } from './pages/Setup'
import { Members } from './pages/Members'
import { MemberDetail } from './pages/MemberDetail'
import { AddMember } from './pages/AddMember'
import { ProfileEdit } from './pages/ProfileEdit'
import { ClassTypes } from './pages/ClassTypes'
import { Payments } from './pages/Payments'
import { MyPayments } from './pages/MyPayments'
import { Classes } from './pages/Classes'
import { Attendance } from './pages/Attendance'
import { MyAttendance } from './pages/MyAttendance'

function App() {
  return (
    <ErrorBoundary>
    <BrowserRouter>
      <AuthProvider>
        <ToastProvider>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/setup" element={<Setup />} />
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route index element={<Dashboard />} />
            <Route path="members" element={<Members />} />
            <Route path="members/new" element={<AddMember />} />
            <Route path="members/:id" element={<MemberDetail />} />
            <Route path="profile" element={<ProfileEdit />} />
            <Route path="classes" element={<Classes />} />
            <Route path="class-types" element={<ClassTypes />} />
            <Route path="payments" element={<Payments />} />
            <Route path="my-payments" element={<MyPayments />} />
            <Route path="attendance" element={<Attendance />} />
            <Route path="my-attendance" element={<MyAttendance />} />
          </Route>
        </Routes>
        </ToastProvider>
      </AuthProvider>
    </BrowserRouter>
    </ErrorBoundary>
  )
}

export default App
