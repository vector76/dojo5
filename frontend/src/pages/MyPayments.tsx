import { useEffect, useState } from 'react'
import { apiFetch } from '../api'
import { useAuth } from '../auth'
import { BalanceSummary } from '../components/BalanceSummary'

interface Payment {
  id: number
  user_id: number
  amount: number
  date: string
  note?: string
  recorded_by: number
}

interface BalanceData {
  expected_balance: number
  total_paid: number
  balance: number
}

export function MyPayments() {
  const { user: currentUser } = useAuth()
  const [payments, setPayments] = useState<Payment[]>([])
  const [balance, setBalance] = useState<BalanceData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!currentUser) return
    Promise.all([
      apiFetch(`/api/users/${currentUser.id}/payments`),
      apiFetch(`/api/users/${currentUser.id}/balance`),
    ])
      .then(async ([paymentsRes, balanceRes]) => {
        const paymentsData: Payment[] = await paymentsRes.json()
        const balanceData: BalanceData = await balanceRes.json()
        setPayments(paymentsData)
        setBalance(balanceData)
        setLoading(false)
      })
      .catch(() => {
        setError('Failed to load payment history')
        setLoading(false)
      })
  }, [currentUser])

  if (loading) return <div>Loading...</div>
  if (error) return <div role="alert">{error}</div>

  return (
    <div>
      <h1>My Payments</h1>

      {balance && (
        <BalanceSummary
          expected_balance={balance.expected_balance}
          total_paid={balance.total_paid}
          balance={balance.balance}
        />
      )}

      {payments.length > 0 ? (
        <table>
          <thead>
            <tr>
              <th>Date</th>
              <th>Amount</th>
              <th>Note</th>
            </tr>
          </thead>
          <tbody>
            {payments.map(p => (
              <tr key={p.id}>
                <td>{p.date}</td>
                <td>${p.amount.toFixed(2)}</td>
                <td>{p.note ?? '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <p>No payments recorded</p>
      )}
    </div>
  )
}
