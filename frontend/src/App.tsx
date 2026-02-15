import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from './auth'
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

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/setup" element={<Setup />} />
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route index element={<Dashboard />} />
            <Route path="members" element={<Members />} />
            <Route path="members/new" element={<AddMember />} />
            <Route path="members/:id" element={<MemberDetail />} />
            <Route path="profile" element={<ProfileEdit />} />
            <Route path="class-types" element={<ClassTypes />} />
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App
