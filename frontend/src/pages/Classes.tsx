import { useEffect, useState } from 'react'
import { useAuth } from '../auth'
import { apiFetch } from '../api'

interface ClassItem {
  id: number
  class_type_id: number
  instructor_id: number
  start_time: string
  duration_minutes: number
  capacity: number
}

interface ClassType {
  id: number
  name: string
}

interface Member {
  id: number
  name: string
  role: string
}

export function Classes() {
  const { user } = useAuth()
  const isAdmin = user?.role === 'admin'
  const isAdminOrInstructor = user?.role === 'admin' || user?.role === 'instructor'

  const [classes, setClasses] = useState<ClassItem[]>([])
  const [classTypes, setClassTypes] = useState<ClassType[]>([])
  const [instructors, setInstructors] = useState<Member[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Date range filter
  const [fromDate, setFromDate] = useState('')
  const [toDate, setToDate] = useState('')

  // Form state
  const [editingId, setEditingId] = useState<number | null>(null)
  const [classTypeId, setClassTypeId] = useState('')
  const [instructorId, setInstructorId] = useState('')
  const [startTime, setStartTime] = useState('')
  const [durationMinutes, setDurationMinutes] = useState('')
  const [capacity, setCapacity] = useState('')
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    loadData()
  }, [])

  async function loadData() {
    try {
      const [classesRes, typesRes] = await Promise.all([
        apiFetch('/api/classes'),
        apiFetch('/api/class-types'),
      ])
      const [classesData, typesData] = await Promise.all([
        classesRes.json() as Promise<ClassItem[]>,
        typesRes.json() as Promise<ClassType[]>,
      ])
      setClasses(classesData)
      setClassTypes(typesData)

      if (isAdminOrInstructor) {
        const membersRes = await apiFetch('/api/users')
        const membersData = await membersRes.json() as Member[]
        setInstructors(membersData.filter(m => m.role === 'admin' || m.role === 'instructor'))
      }
    } catch {
      setError('Failed to load class schedule')
    } finally {
      setLoading(false)
    }
  }

  async function loadClasses(from?: string, to?: string) {
    try {
      const params = new URLSearchParams()
      if (from) params.set('from', new Date(from).toISOString())
      if (to) params.set('to', new Date(to + 'T23:59:59').toISOString())
      const url = '/api/classes' + (params.toString() ? `?${params}` : '')
      const res = await apiFetch(url)
      const data: ClassItem[] = await res.json()
      setClasses(data)
      setError('')
    } catch {
      setError('Failed to load classes')
    }
  }

  function handleFilter(e: React.FormEvent) {
    e.preventDefault()
    loadClasses(fromDate, toDate)
  }

  function clearFilter() {
    setFromDate('')
    setToDate('')
    loadClasses()
  }

  function getClassTypeName(id: number): string {
    return classTypes.find(ct => ct.id === id)?.name ?? `Type #${id}`
  }

  function getInstructorName(id: number): string {
    return instructors.find(m => m.id === id)?.name ?? `Instructor #${id}`
  }

  function formatDateTime(isoString: string): string {
    const d = new Date(isoString)
    return d.toLocaleString()
  }

  function startCreate() {
    setClassTypeId(classTypes[0]?.id.toString() ?? '')
    setInstructorId(instructors[0]?.id.toString() ?? '')
    setStartTime('')
    setDurationMinutes('60')
    setCapacity('20')
    setFormError('')
    setEditingId(-1)
  }

  function startEdit(cls: ClassItem) {
    setEditingId(cls.id)
    setClassTypeId(cls.class_type_id.toString())
    setInstructorId(cls.instructor_id.toString())
    // Convert ISO to datetime-local format
    const dt = new Date(cls.start_time)
    const local = dt.getFullYear() + '-' +
      String(dt.getMonth() + 1).padStart(2, '0') + '-' +
      String(dt.getDate()).padStart(2, '0') + 'T' +
      String(dt.getHours()).padStart(2, '0') + ':' +
      String(dt.getMinutes()).padStart(2, '0')
    setStartTime(local)
    setDurationMinutes(cls.duration_minutes.toString())
    setCapacity(cls.capacity.toString())
    setFormError('')
  }

  function cancelForm() {
    setEditingId(null)
    setFormError('')
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setFormError('')
    try {
      const body = {
        class_type_id: Number(classTypeId),
        instructor_id: Number(instructorId),
        start_time: new Date(startTime).toISOString(),
        duration_minutes: Number(durationMinutes),
        capacity: Number(capacity),
      }
      if (editingId === -1) {
        const res = await apiFetch('/api/classes', {
          method: 'POST',
          body: JSON.stringify(body),
        })
        const created: ClassItem = await res.json()
        setClasses(prev => [...prev, created].sort((a, b) => a.start_time.localeCompare(b.start_time)))
      } else {
        const res = await apiFetch(`/api/classes/${editingId}`, {
          method: 'PUT',
          body: JSON.stringify(body),
        })
        const updated: ClassItem = await res.json()
        setClasses(prev => prev.map(c => c.id === updated.id ? updated : c).sort((a, b) => a.start_time.localeCompare(b.start_time)))
      }
      cancelForm()
    } catch {
      setFormError(editingId === -1 ? 'Failed to create class' : 'Failed to update class')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Are you sure you want to delete this class?')) return
    try {
      await apiFetch(`/api/classes/${id}`, { method: 'DELETE' })
      setClasses(prev => prev.filter(c => c.id !== id))
      setError('')
    } catch {
      setError('Failed to delete class')
    }
  }

  if (loading) return <div>Loading...</div>
  if (error && classes.length === 0) return <div role="alert">{error}</div>

  return (
    <div>
      <h1>Class Schedule</h1>

      <form onSubmit={handleFilter} aria-label="Filter classes" className="filters">
        <label htmlFor="from-date">From
          <input id="from-date" type="date" value={fromDate} onChange={e => setFromDate(e.target.value)} />
        </label>
        <label htmlFor="to-date">To
          <input id="to-date" type="date" value={toDate} onChange={e => setToDate(e.target.value)} />
        </label>
        <button type="submit">Filter</button>
        <button type="button" onClick={clearFilter}>Clear</button>
      </form>

      {error && <div role="alert">{error}</div>}

      {isAdmin && <button onClick={startCreate}>Add Class</button>}

      {editingId !== null && (
        <form onSubmit={handleSubmit} aria-label={editingId === -1 ? 'Create class' : 'Edit class'}>
          <div>
            <label htmlFor="class-type">Class Type</label>
            <select id="class-type" value={classTypeId} onChange={e => setClassTypeId(e.target.value)} required>
              {classTypes.map(ct => (
                <option key={ct.id} value={ct.id}>{ct.name}</option>
              ))}
            </select>
          </div>
          <div>
            <label htmlFor="instructor">Instructor</label>
            <select id="instructor" value={instructorId} onChange={e => setInstructorId(e.target.value)} required>
              {instructors.map(m => (
                <option key={m.id} value={m.id}>{m.name}</option>
              ))}
            </select>
          </div>
          <div>
            <label htmlFor="start-time">Start Time</label>
            <input id="start-time" type="datetime-local" value={startTime} onChange={e => setStartTime(e.target.value)} required />
          </div>
          <div>
            <label htmlFor="duration">Duration (minutes)</label>
            <input id="duration" type="number" value={durationMinutes} onChange={e => setDurationMinutes(e.target.value)} min="1" required />
          </div>
          <div>
            <label htmlFor="capacity">Capacity</label>
            <input id="capacity" type="number" value={capacity} onChange={e => setCapacity(e.target.value)} min="1" required />
          </div>
          {formError && <div role="alert">{formError}</div>}
          <button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save'}</button>
          <button type="button" onClick={cancelForm}>Cancel</button>
        </form>
      )}

      <table>
        <thead>
          <tr>
            <th>Date/Time</th>
            <th>Class Type</th>
            <th>Instructor</th>
            <th>Duration</th>
            <th>Capacity</th>
            {isAdmin && <th>Actions</th>}
          </tr>
        </thead>
        <tbody>
          {classes.map(cls => (
            <tr key={cls.id}>
              <td data-label="Date/Time">{formatDateTime(cls.start_time)}</td>
              <td data-label="Class Type">{getClassTypeName(cls.class_type_id)}</td>
              <td data-label="Instructor">{getInstructorName(cls.instructor_id)}</td>
              <td data-label="Duration">{cls.duration_minutes} min</td>
              <td data-label="Capacity">{cls.capacity}</td>
              {isAdmin && (
                <td data-label="Actions">
                  <button onClick={() => startEdit(cls)}>Edit</button>
                  <button onClick={() => handleDelete(cls.id)}>Delete</button>
                </td>
              )}
            </tr>
          ))}
          {classes.length === 0 && (
            <tr>
              <td colSpan={isAdmin ? 6 : 5}>No classes found</td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  )
}
