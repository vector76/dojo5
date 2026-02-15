import { useEffect, useState } from 'react'
import { apiFetch } from '../api'

interface ClassType {
  id: number
  name: string
  description?: string
}

export function ClassTypes() {
  const [classTypes, setClassTypes] = useState<ClassType[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Form state
  const [editingId, setEditingId] = useState<number | null>(null)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    loadClassTypes()
  }, [])

  async function loadClassTypes() {
    try {
      const res = await apiFetch('/api/class-types')
      const data: ClassType[] = await res.json()
      setClassTypes(data)
    } catch {
      setError('Failed to load class types')
    } finally {
      setLoading(false)
    }
  }

  function startCreate() {
    setName('')
    setDescription('')
    setFormError('')
    setEditingId(-1) // -1 signals "creating new"
  }

  function startEdit(ct: ClassType) {
    setEditingId(ct.id)
    setName(ct.name)
    setDescription(ct.description ?? '')
    setFormError('')
  }

  function cancelForm() {
    setEditingId(null)
    setName('')
    setDescription('')
    setFormError('')
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setFormError('')
    try {
      if (editingId === -1) {
        // Create
        const res = await apiFetch('/api/class-types', {
          method: 'POST',
          body: JSON.stringify({ name, description: description || null }),
        })
        const created: ClassType = await res.json()
        setClassTypes(prev => [...prev, created])
      } else {
        // Update
        const res = await apiFetch(`/api/class-types/${editingId}`, {
          method: 'PUT',
          body: JSON.stringify({ name, description: description || null }),
        })
        const updated: ClassType = await res.json()
        setClassTypes(prev => prev.map(ct => ct.id === updated.id ? updated : ct))
      }
      cancelForm()
    } catch {
      setFormError(editingId === -1 ? 'Failed to create class type' : 'Failed to update class type')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Are you sure you want to delete this class type?')) return
    try {
      await apiFetch(`/api/class-types/${id}`, { method: 'DELETE' })
      setClassTypes(prev => prev.filter(ct => ct.id !== id))
    } catch {
      setError('Failed to delete class type')
    }
  }

  if (loading) return <div>Loading...</div>
  if (error) return <div role="alert">{error}</div>

  return (
    <div>
      <h1>Class Types</h1>
      <button onClick={startCreate}>Add Class Type</button>

      {editingId !== null && (
        <form onSubmit={handleSubmit} aria-label={editingId === -1 ? 'Create class type' : 'Edit class type'}>
          <div>
            <label htmlFor="name">Name</label>
            <input id="name" value={name} onChange={e => setName(e.target.value)} required />
          </div>
          <div>
            <label htmlFor="description">Description</label>
            <textarea id="description" value={description} onChange={e => setDescription(e.target.value)} />
          </div>
          {formError && <div role="alert">{formError}</div>}
          <button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save'}</button>
          <button type="button" onClick={cancelForm}>Cancel</button>
        </form>
      )}

      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Description</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {classTypes.map(ct => (
            <tr key={ct.id}>
              <td>{ct.name}</td>
              <td>{ct.description ?? '-'}</td>
              <td>
                <button onClick={() => startEdit(ct)}>Edit</button>
                <button onClick={() => handleDelete(ct.id)}>Delete</button>
              </td>
            </tr>
          ))}
          {classTypes.length === 0 && (
            <tr>
              <td colSpan={3}>No class types found</td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  )
}
