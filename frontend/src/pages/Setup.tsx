import { useState } from 'react'
import { apiFetch, ApiError, setToken } from '../api'

export function Setup() {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      const res = await apiFetch('/api/auth/setup', {
        method: 'POST',
        body: JSON.stringify({ name, email, phone, password }),
      })
      const data: { token: string } = await res.json()
      setToken(data.token)
      // Reload so AuthProvider picks up the new token
      window.location.href = '/'
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 409) {
          setError('Setup has already been completed')
        } else {
          setError(err.message)
        }
      } else {
        setError('An unexpected error occurred')
      }
    }
  }

  return (
    <div className="setup-page">
      <h1>Welcome to Dojo CRM</h1>
      <p>Create the initial administrator account to get started.</p>
      <form onSubmit={handleSubmit}>
        {error && <p className="error" role="alert">{error}</p>}
        <label>
          Name
          <input
            type="text"
            value={name}
            onChange={e => setName(e.target.value)}
            required
          />
        </label>
        <label>
          Email
          <input
            type="email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            required
          />
        </label>
        <label>
          Phone
          <input
            type="tel"
            value={phone}
            onChange={e => setPhone(e.target.value)}
            required
          />
        </label>
        <label>
          Password
          <input
            type="password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            required
          />
        </label>
        <button type="submit">Create Admin Account</button>
      </form>
    </div>
  )
}
