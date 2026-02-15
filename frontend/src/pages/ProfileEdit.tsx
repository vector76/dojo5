import { useEffect, useState } from 'react'
import { apiFetch } from '../api'
import { useAuth } from '../auth'

export function ProfileEdit() {
  const { user: currentUser } = useAuth()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [emergencyContact, setEmergencyContact] = useState('')
  const [saveError, setSaveError] = useState('')
  const [saving, setSaving] = useState(false)
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    if (!currentUser) return
    apiFetch(`/api/users/${currentUser.id}`)
      .then(res => res.json())
      .then((data: { name: string; email: string; phone: string; emergency_contact?: string }) => {
        setName(data.name)
        setEmail(data.email)
        setPhone(data.phone)
        setEmergencyContact(data.emergency_contact ?? '')
        setLoading(false)
      })
      .catch(() => {
        setError('Failed to load profile')
        setLoading(false)
      })
  }, [currentUser])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!currentUser) return
    setSaving(true)
    setSaveError('')
    setSuccess(false)
    try {
      await apiFetch(`/api/users/${currentUser.id}`, {
        method: 'PUT',
        body: JSON.stringify({
          name,
          email,
          phone,
          emergency_contact: emergencyContact || null,
        }),
      })
      setSuccess(true)
    } catch {
      setSaveError('Failed to save profile')
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <div>Loading...</div>
  if (error) return <div role="alert">{error}</div>

  return (
    <div>
      <h1>Edit Profile</h1>
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
          <label htmlFor="emergencyContact">Emergency Contact</label>
          <input id="emergencyContact" value={emergencyContact} onChange={e => setEmergencyContact(e.target.value)} />
        </div>
        {saveError && <div role="alert">{saveError}</div>}
        {success && <div>Profile updated successfully</div>}
        <button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save'}</button>
      </form>
    </div>
  )
}
