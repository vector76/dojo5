import { useEffect, useState } from 'react'
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
}

interface AttendanceRecord {
  id: number
  class_id: number
  user_id: number
  checked_in_at: string
}

export function Attendance() {
  const [classes, setClasses] = useState<ClassItem[]>([])
  const [classTypes, setClassTypes] = useState<ClassType[]>([])
  const [members, setMembers] = useState<Member[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Selected class
  const [selectedClassId, setSelectedClassId] = useState('')
  const [attendance, setAttendance] = useState<AttendanceRecord[]>([])
  const [loadingAttendance, setLoadingAttendance] = useState(false)

  // Check-in
  const [checkInUserId, setCheckInUserId] = useState('')
  const [checkInError, setCheckInError] = useState('')

  useEffect(() => {
    loadData()
  }, [])

  async function loadData() {
    try {
      const [classesRes, typesRes, membersRes] = await Promise.all([
        apiFetch('/api/classes'),
        apiFetch('/api/class-types'),
        apiFetch('/api/users'),
      ])
      const [classesData, typesData, membersData] = await Promise.all([
        classesRes.json() as Promise<ClassItem[]>,
        typesRes.json() as Promise<ClassType[]>,
        membersRes.json() as Promise<Member[]>,
      ])
      setClasses(classesData)
      setClassTypes(typesData)
      setMembers(membersData)
    } catch {
      setError('Failed to load data')
    } finally {
      setLoading(false)
    }
  }

  async function loadAttendance(classId: string) {
    setLoadingAttendance(true)
    setCheckInError('')
    setError('')
    try {
      const res = await apiFetch(`/api/classes/${classId}/attendance`)
      const data: AttendanceRecord[] = await res.json()
      setAttendance(data)
    } catch {
      setError('Failed to load attendance')
    } finally {
      setLoadingAttendance(false)
    }
  }

  function handleClassSelect(classId: string) {
    setSelectedClassId(classId)
    setAttendance([])
    setCheckInError('')
    if (classId) {
      loadAttendance(classId)
    }
  }

  function getClassTypeName(id: number): string {
    return classTypes.find(ct => ct.id === id)?.name ?? `Type #${id}`
  }

  function formatClassLabel(cls: ClassItem): string {
    const dt = new Date(cls.start_time)
    const typeName = getClassTypeName(cls.class_type_id)
    return `${typeName} — ${dt.toLocaleString()}`
  }

  function getMemberName(id: number): string {
    return members.find(m => m.id === id)?.name ?? `User #${id}`
  }

  function getCheckedInUserIds(): Set<number> {
    return new Set(attendance.map(a => a.user_id))
  }

  async function handleCheckIn(e: React.FormEvent) {
    e.preventDefault()
    setCheckInError('')
    try {
      const res = await apiFetch(`/api/classes/${selectedClassId}/attendance`, {
        method: 'POST',
        body: JSON.stringify({ user_id: Number(checkInUserId) }),
      })
      const record: AttendanceRecord = await res.json()
      setAttendance(prev => [...prev, record])
      setCheckInUserId('')
    } catch {
      setCheckInError('Failed to check in member')
    }
  }

  if (loading) return <div>Loading...</div>
  if (error && classes.length === 0) return <div role="alert">{error}</div>

  const checkedInIds = getCheckedInUserIds()
  const availableMembers = members.filter(m => !checkedInIds.has(m.id))

  return (
    <div>
      <h1>Attendance</h1>

      <div>
        <label htmlFor="class-select">Select Class</label>
        <select id="class-select" value={selectedClassId} onChange={e => handleClassSelect(e.target.value)}>
          <option value="">-- Select a class --</option>
          {classes.map(cls => (
            <option key={cls.id} value={cls.id}>{formatClassLabel(cls)}</option>
          ))}
        </select>
      </div>

      {error && <div role="alert">{error}</div>}

      {selectedClassId && (
        <>
          <section aria-label="Check In">
            <h2>Check In</h2>
            {availableMembers.length > 0 ? (
              <form onSubmit={handleCheckIn} aria-label="Check in member">
                <label htmlFor="member-select">Member</label>
                <select id="member-select" value={checkInUserId} onChange={e => setCheckInUserId(e.target.value)} required>
                  <option value="">-- Select member --</option>
                  {availableMembers.map(m => (
                    <option key={m.id} value={m.id}>{m.name}</option>
                  ))}
                </select>
                <button type="submit">Check In</button>
              </form>
            ) : (
              <p>All members are checked in.</p>
            )}
            {checkInError && <div role="alert">{checkInError}</div>}
          </section>

          <section aria-label="Checked In">
            <h2>Checked In</h2>
            {loadingAttendance ? (
              <div>Loading attendance...</div>
            ) : attendance.length === 0 ? (
              <p>No one checked in yet.</p>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>Member</th>
                    <th>Checked In At</th>
                  </tr>
                </thead>
                <tbody>
                  {attendance.map(a => (
                    <tr key={a.id}>
                      <td data-label="Member">{getMemberName(a.user_id)}</td>
                      <td data-label="Checked In At">{new Date(a.checked_in_at).toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </section>
        </>
      )}
    </div>
  )
}
