import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { apiFetch } from '../api'
import { useAuth } from '../auth'

interface MemberData {
  id: number
  name: string
  email: string
  phone: string
  role: string
  membership_type?: string
  membership_status?: string
  emergency_contact?: string
  join_date?: string
}

interface BalanceData {
  expected_balance: number
  total_paid: number
  balance: number
}

export function MemberDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user: currentUser } = useAuth()
  const isAdmin = currentUser?.role === 'admin'

  const [member, setMember] = useState<MemberData | null>(null)
  const [balance, setBalance] = useState<BalanceData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Edit form state
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [membershipType, setMembershipType] = useState('')
  const [membershipStatus, setMembershipStatus] = useState('')
  const [emergencyContact, setEmergencyContact] = useState('')
  const [joinDate, setJoinDate] = useState('')
  const [role, setRole] = useState('')
  const [saveError, setSaveError] = useState('')
  const [saving, setSaving] = useState(false)

  // Password reset state
  const [newPassword, setNewPassword] = useState('')
  const [passwordMsg, setPasswordMsg] = useState('')

  // Balance state
  const [expectedBalance, setExpectedBalance] = useState('')
  const [balanceMsg, setBalanceMsg] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const res = await apiFetch(`/api/users/${id}`)
        const data: MemberData = await res.json()
        setMember(data)
        populateForm(data)

        if (currentUser?.role === 'admin') {
          const balRes = await apiFetch(`/api/users/${id}/balance`)
          const balData: BalanceData = await balRes.json()
          setBalance(balData)
          setExpectedBalance(String(balData.expected_balance))
        }
      } catch {
        setError('Failed to load member')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [id, currentUser?.role])

  function populateForm(data: MemberData) {
    setName(data.name)
    setEmail(data.email)
    setPhone(data.phone)
    setMembershipType(data.membership_type ?? '')
    setMembershipStatus(data.membership_status ?? '')
    setEmergencyContact(data.emergency_contact ?? '')
    setJoinDate(data.join_date ?? '')
    setRole(data.role)
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setSaveError('')
    try {
      const body: Record<string, string | null> = {
        name,
        email,
        phone,
        emergency_contact: emergencyContact || null,
      }
      if (isAdmin) {
        body.membership_type = membershipType || null
        body.membership_status = membershipStatus || null
        body.join_date = joinDate || null
        body.role = role
      }
      const res = await apiFetch(`/api/users/${id}`, {
        method: 'PUT',
        body: JSON.stringify(body),
      })
      const updated: MemberData = await res.json()
      setMember(updated)
      populateForm(updated)
      setEditing(false)
    } catch {
      setSaveError('Failed to save changes')
    } finally {
      setSaving(false)
    }
  }

  async function handlePasswordReset(e: React.FormEvent) {
    e.preventDefault()
    setPasswordMsg('')
    try {
      await apiFetch(`/api/users/${id}/password`, {
        method: 'PUT',
        body: JSON.stringify({ password: newPassword }),
      })
      setNewPassword('')
      setPasswordMsg('Password updated')
    } catch {
      setPasswordMsg('Failed to reset password')
    }
  }

  async function handleSetBalance(e: React.FormEvent) {
    e.preventDefault()
    setBalanceMsg('')
    try {
      const res = await apiFetch(`/api/users/${id}/balance`, {
        method: 'PUT',
        body: JSON.stringify({ expected_balance: Number(expectedBalance) }),
      })
      const balData: BalanceData = await res.json()
      setBalance(balData)
      setExpectedBalance(String(balData.expected_balance))
      setBalanceMsg('Balance updated')
    } catch {
      setBalanceMsg('Failed to update balance')
    }
  }

  async function handleDelete() {
    if (!confirm('Are you sure you want to delete this member?')) return
    try {
      await apiFetch(`/api/users/${id}`, { method: 'DELETE' })
      navigate('/members')
    } catch {
      setSaveError('Failed to delete member')
    }
  }

  if (loading) return <div>Loading...</div>
  if (error) return <div role="alert">{error}</div>
  if (!member) return null

  return (
    <div>
      <div>
        <Link to="/members">&larr; Back to Members</Link>
      </div>
      <h1>{member.name}</h1>

      {!editing ? (
        <div>
          <dl>
            <dt>Email</dt>
            <dd>{member.email}</dd>
            <dt>Phone</dt>
            <dd>{member.phone}</dd>
            <dt>Role</dt>
            <dd>{member.role}</dd>
            <dt>Membership Type</dt>
            <dd>{member.membership_type ?? '-'}</dd>
            <dt>Membership Status</dt>
            <dd>{member.membership_status ?? '-'}</dd>
            <dt>Emergency Contact</dt>
            <dd>{member.emergency_contact ?? '-'}</dd>
            <dt>Join Date</dt>
            <dd>{member.join_date ?? '-'}</dd>
          </dl>
          <button onClick={() => setEditing(true)}>Edit</button>
          {isAdmin && (
            <button onClick={handleDelete}>Delete Member</button>
          )}
        </div>
      ) : (
        <form onSubmit={handleSave}>
          <div>
            <label htmlFor="name">Name</label>
            <input id="name" value={name} onChange={e => setName(e.target.value)} required />
          </div>
          <div>
            <label htmlFor="email">Email</label>
            <input id="email" type="email" value={email} onChange={e => setEmail(e.target.value)} required />
          </div>
          <div>
            <label htmlFor="phone">Phone</label>
            <input id="phone" value={phone} onChange={e => setPhone(e.target.value)} required />
          </div>
          <div>
            <label htmlFor="emergencyContact">Emergency Contact</label>
            <input id="emergencyContact" value={emergencyContact} onChange={e => setEmergencyContact(e.target.value)} />
          </div>
          {isAdmin && (
            <>
              <div>
                <label htmlFor="role">Role</label>
                <select id="role" value={role} onChange={e => setRole(e.target.value)}>
                  <option value="user">user</option>
                  <option value="instructor">instructor</option>
                  <option value="admin">admin</option>
                </select>
              </div>
              <div>
                <label htmlFor="membershipType">Membership Type</label>
                <input id="membershipType" value={membershipType} onChange={e => setMembershipType(e.target.value)} />
              </div>
              <div>
                <label htmlFor="membershipStatus">Membership Status</label>
                <input id="membershipStatus" value={membershipStatus} onChange={e => setMembershipStatus(e.target.value)} />
              </div>
              <div>
                <label htmlFor="joinDate">Join Date</label>
                <input id="joinDate" type="date" value={joinDate} onChange={e => setJoinDate(e.target.value)} />
              </div>
            </>
          )}
          {saveError && <div role="alert">{saveError}</div>}
          <button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save'}</button>
          <button type="button" onClick={() => { populateForm(member); setEditing(false); setSaveError('') }}>Cancel</button>
        </form>
      )}

      {isAdmin && (
        <>
          <section aria-label="Password Reset">
            <h2>Reset Password</h2>
            <form onSubmit={handlePasswordReset}>
              <div>
                <label htmlFor="newPassword">New Password</label>
                <input id="newPassword" type="password" value={newPassword} onChange={e => setNewPassword(e.target.value)} required />
              </div>
              {passwordMsg && <div>{passwordMsg}</div>}
              <button type="submit">Reset Password</button>
            </form>
          </section>

          <section aria-label="Balance Management">
            <h2>Balance</h2>
            {balance && (
              <dl>
                <dt>Expected Balance</dt>
                <dd>${balance.expected_balance.toFixed(2)}</dd>
                <dt>Total Paid</dt>
                <dd>${balance.total_paid.toFixed(2)}</dd>
                <dt>Outstanding Balance</dt>
                <dd>${balance.balance.toFixed(2)}</dd>
              </dl>
            )}
            <form onSubmit={handleSetBalance}>
              <div>
                <label htmlFor="expectedBalance">Set Expected Balance</label>
                <input id="expectedBalance" type="number" step="0.01" value={expectedBalance} onChange={e => setExpectedBalance(e.target.value)} required />
              </div>
              {balanceMsg && <div>{balanceMsg}</div>}
              <button type="submit">Update Balance</button>
            </form>
          </section>
        </>
      )}
    </div>
  )
}
