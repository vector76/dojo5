import { useEffect, useState } from 'react'
import { useAuth } from '../auth'
import { apiFetch } from '../api'
import { BalanceSummary } from '../components/BalanceSummary'

interface BalanceData {
  expected_balance: number
  total_paid: number
  balance: number
}

export function Dashboard() {
  const { user } = useAuth()
  const [balance, setBalance] = useState<BalanceData | null>(null)

  useEffect(() => {
    if (!user) return
    apiFetch(`/api/users/${user.id}/balance`)
      .then(res => res.json())
      .then((data: BalanceData) => setBalance(data))
      .catch(() => {
        // Balance display is optional on dashboard; silently ignore errors
      })
  }, [user])

  return (
    <div>
      <h1>Dashboard</h1>
      <p>Welcome, {user?.name}!</p>
      {balance && (
        <section aria-label="My Balance">
          <h2>My Balance</h2>
          <BalanceSummary
            expected_balance={balance.expected_balance}
            total_paid={balance.total_paid}
            balance={balance.balance}
          />
        </section>
      )}
    </div>
  )
}
