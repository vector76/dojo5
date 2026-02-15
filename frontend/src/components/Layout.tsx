import { useState, useEffect } from 'react'
import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../auth'

export function Layout() {
  const { user, logout } = useAuth()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const location = useLocation()

  // Close sidebar on navigation (mobile)
  useEffect(() => {
    setSidebarOpen(false)
  }, [location.pathname])

  if (!user) return null

  const isAdmin = user.role === 'admin'
  const isAdminOrInstructor = user.role === 'admin' || user.role === 'instructor'

  return (
    <div className="layout">
      <button
        className="menu-toggle"
        onClick={() => setSidebarOpen(!sidebarOpen)}
        aria-label={sidebarOpen ? 'Close menu' : 'Open menu'}
      >
        {sidebarOpen ? '\u2715' : '\u2630'}
      </button>
      {sidebarOpen && (
        <div className="sidebar-overlay" onClick={() => setSidebarOpen(false)} />
      )}
      <nav className={`sidebar${sidebarOpen ? ' open' : ''}`} role="navigation">
        <div className="sidebar-header">
          <h2>Dojo CRM</h2>
        </div>
        <ul>
          <li><NavLink to="/">Dashboard</NavLink></li>
          <li><NavLink to="/profile">Profile</NavLink></li>
          <li><NavLink to="/classes">Classes</NavLink></li>
          <li><NavLink to="/my-payments">My Payments</NavLink></li>
          <li><NavLink to="/my-attendance">My Attendance</NavLink></li>
          {isAdminOrInstructor && (
            <li><NavLink to="/members">Members</NavLink></li>
          )}
          {isAdminOrInstructor && (
            <li><NavLink to="/attendance">Attendance</NavLink></li>
          )}
          {isAdmin && (
            <li><NavLink to="/payments">Payments</NavLink></li>
          )}
          {isAdmin && (
            <li><NavLink to="/class-types">Class Types</NavLink></li>
          )}
        </ul>
        <div className="sidebar-footer">
          <span>{user.name}</span>
          <button onClick={logout}>Logout</button>
        </div>
      </nav>
      <main className="content">
        <Outlet />
      </main>
    </div>
  )
}
