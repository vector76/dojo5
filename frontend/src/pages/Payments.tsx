import { useEffect, useState } from 'react'
import { apiFetch } from '../api'

interface Member {
  id: number
  name: string
  email: string
}

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

export function Payments() {
  const [members, setMembers] = useState<Member[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Record payment form
  const [selectedUserId, setSelectedUserId] = useState('')
  const [amount, setAmount] = useState('')
  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10))
  const [note, setNote] = useState('')
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)
  const [formSuccess, setFormSuccess] = useState('')

  // History view
  const [historyUserId, setHistoryUserId] = useState('')
  const [payments, setPayments] = useState<Payment[]>([])
  const [balance, setBalance] = useState<BalanceData | null>(null)
  const [historyLoading, setHistoryLoading] = useState(false)
  const [historyError, setHistoryError] = useState('')

  useEffect(() => {
    apiFetch('/api/users')
      .then(res => res.json())
      .then((data: Member[]) => {
        setMembers(data)
        setLoading(false)
      })
      .catch(() => {
        setError('Failed to load members')
        setLoading(false)
      })
  }, [])

  async function handleRecordPayment(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setFormError('')
    setFormSuccess('')
    try {
      await apiFetch('/api/payments', {
        method: 'POST',
        body: JSON.stringify({
          user_id: Number(selectedUserId),
          amount: Number(amount),
          date,
          note: note || null,
        }),
      })
      setAmount('')
      setNote('')
      setFormSuccess('Payment recorded')
      // Refresh history if viewing the same user
      if (historyUserId === selectedUserId) {
        loadHistory(selectedUserId)
      }
    } catch {
      setFormError('Failed to record payment')
    } finally {
      setSaving(false)
    }
  }

  async function loadHistory(userId: string) {
    setHistoryLoading(true)
    setHistoryError('')
    try {
      const [paymentsRes, balanceRes] = await Promise.all([
        apiFetch(`/api/users/${userId}/payments`),
        apiFetch(`/api/users/${userId}/balance`),
      ])
      const paymentsData: Payment[] = await paymentsRes.json()
      const balanceData: BalanceData = await balanceRes.json()
      setPayments(paymentsData)
      setBalance(balanceData)
    } catch {
      setHistoryError('Failed to load payment history')
    } finally {
      setHistoryLoading(false)
    }
  }

  function handleViewHistory(e: React.FormEvent) {
    e.preventDefault()
    if (historyUserId) {
      loadHistory(historyUserId)
    }
  }

  if (loading) return <div>Loading...</div>
  if (error) return <div role="alert">{error}</div>

  return (
    <div>
      <h1>Payments</h1>

      <section aria-label="Record Payment">
        <h2>Record Payment</h2>
        <form onSubmit={handleRecordPayment}>
          <div>
            <label htmlFor="paymentMember">Member</label>
            <select id="paymentMember" value={selectedUserId} onChange={e => setSelectedUserId(e.target.value)} required>
              <option value="">Select a member</option>
              {members.map(m => (
                <option key={m.id} value={m.id}>{m.name} ({m.email})</option>
              ))}
            </select>
          </div>
          <div>
            <label htmlFor="amount">Amount</label>
            <input id="amount" type="number" step="0.01" value={amount} onChange={e => setAmount(e.target.value)} required />
          </div>
          <div>
            <label htmlFor="date">Date</label>
            <input id="date" type="date" value={date} onChange={e => setDate(e.target.value)} required />
          </div>
          <div>
            <label htmlFor="note">Note</label>
            <input id="note" value={note} onChange={e => setNote(e.target.value)} />
          </div>
          {formError && <div role="alert">{formError}</div>}
          {formSuccess && <div>{formSuccess}</div>}
          <button type="submit" disabled={saving}>{saving ? 'Recording...' : 'Record Payment'}</button>
        </form>
      </section>

      <section aria-label="Payment History">
        <h2>Payment History</h2>
        <form onSubmit={handleViewHistory}>
          <div>
            <label htmlFor="historyMember">Member</label>
            <select id="historyMember" value={historyUserId} onChange={e => setHistoryUserId(e.target.value)} required>
              <option value="">Select a member</option>
              {members.map(m => (
                <option key={m.id} value={m.id}>{m.name} ({m.email})</option>
              ))}
            </select>
          </div>
          <button type="submit">View History</button>
        </form>

        {historyLoading && <div>Loading history...</div>}
        {historyError && <div role="alert">{historyError}</div>}

        {balance && !historyLoading && (
          <dl>
            <dt>Expected Balance</dt>
            <dd>${balance.expected_balance.toFixed(2)}</dd>
            <dt>Total Paid</dt>
            <dd>${balance.total_paid.toFixed(2)}</dd>
            <dt>Outstanding Balance</dt>
            <dd>${balance.balance.toFixed(2)}</dd>
          </dl>
        )}

        {payments.length > 0 && !historyLoading && (
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
        )}

        {payments.length === 0 && balance && !historyLoading && (
          <p>No payments recorded</p>
        )}
      </section>
    </div>
  )
}
