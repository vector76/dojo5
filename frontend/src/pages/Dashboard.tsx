import { useAuth } from '../auth'

export function Dashboard() {
  const { user } = useAuth()

  return (
    <div>
      <h1>Dashboard</h1>
      <p>Welcome, {user?.name}!</p>
    </div>
  )
}
