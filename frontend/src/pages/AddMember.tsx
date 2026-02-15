import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { apiFetch } from '../api'

export function AddMember() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState('user')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setError('')
    try {
      await apiFetch('/api/users', {
        method: 'POST',
        body: JSON.stringify({ name, email, phone, password, role }),
      })
      navigate('/members')
    } catch {
      setError('Failed to create member')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <div>
        <Link to="/members">&larr; Back to Members</Link>
      </div>
      <h1>Add Member</h1>
      <form onSubmit={handleSubmit}>
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
          <label htmlFor="password">Password</label>
          <input id="password" type="password" value={password} onChange={e => setPassword(e.target.value)} required />
        </div>
        <div>
          <label htmlFor="role">Role</label>
          <select id="role" value={role} onChange={e => setRole(e.target.value)}>
            <option value="user">user</option>
            <option value="instructor">instructor</option>
            <option value="admin">admin</option>
          </select>
        </div>
        {error && <div role="alert">{error}</div>}
        <button type="submit" disabled={saving}>{saving ? 'Creating...' : 'Create Member'}</button>
      </form>
    </div>
  )
}
