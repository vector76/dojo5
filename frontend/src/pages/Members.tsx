import { useEffect, useState } from 'react'
import { apiFetch } from '../api'

interface Member {
  id: number
  name: string
  email: string
  phone: string
  role: string
  membership_type?: string
  membership_status?: string
}

export function Members() {
  const [members, setMembers] = useState<Member[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [roleFilter, setRoleFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState('')

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

  if (loading) return <div>Loading...</div>
  if (error) return <div role="alert">{error}</div>

  const filtered = members.filter(m => {
    const matchesSearch =
      search === '' ||
      m.name.toLowerCase().includes(search.toLowerCase()) ||
      m.email.toLowerCase().includes(search.toLowerCase())

    const matchesRole = roleFilter === '' || m.role === roleFilter

    const matchesStatus =
      statusFilter === '' || (m.membership_status ?? '') === statusFilter

    return matchesSearch && matchesRole && matchesStatus
  })

  const roles = [...new Set(members.map(m => m.role))].sort()
  const statuses = [
    ...new Set(members.map(m => m.membership_status).filter(Boolean)),
  ].sort() as string[]

  return (
    <div>
      <h1>Members</h1>
      <div className="filters">
        <input
          type="text"
          placeholder="Search by name or email"
          aria-label="Search members"
          value={search}
          onChange={e => setSearch(e.target.value)}
        />
        <select
          aria-label="Filter by role"
          value={roleFilter}
          onChange={e => setRoleFilter(e.target.value)}
        >
          <option value="">All roles</option>
          {roles.map(r => (
            <option key={r} value={r}>{r}</option>
          ))}
        </select>
        <select
          aria-label="Filter by status"
          value={statusFilter}
          onChange={e => setStatusFilter(e.target.value)}
        >
          <option value="">All statuses</option>
          {statuses.map(s => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </div>
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Email</th>
            <th>Phone</th>
            <th>Role</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          {filtered.map(m => (
            <tr key={m.id}>
              <td>{m.name}</td>
              <td>{m.email}</td>
              <td>{m.phone}</td>
              <td>{m.role}</td>
              <td>{m.membership_status ?? '-'}</td>
            </tr>
          ))}
          {filtered.length === 0 && (
            <tr>
              <td colSpan={5}>No members found</td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  )
}
