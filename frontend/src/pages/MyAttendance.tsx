import { useEffect, useState } from 'react'
import { useAuth } from '../auth'
import { apiFetch } from '../api'

interface AttendanceRecord {
  id: number
  class_id: number
  user_id: number
  checked_in_at: string
}

interface ClassItem {
  id: number
  class_type_id: number
  start_time: string
  duration_minutes: number
}

interface ClassType {
  id: number
  name: string
}

export function MyAttendance() {
  const { user } = useAuth()
  const [attendance, setAttendance] = useState<AttendanceRecord[]>([])
  const [classes, setClasses] = useState<ClassItem[]>([])
  const [classTypes, setClassTypes] = useState<ClassType[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (user) loadData()
  }, [user])

  async function loadData() {
    try {
      const [attendanceRes, classesRes, typesRes] = await Promise.all([
        apiFetch(`/api/users/${user!.id}/attendance`),
        apiFetch('/api/classes'),
        apiFetch('/api/class-types'),
      ])
      const [attendanceData, classesData, typesData] = await Promise.all([
        attendanceRes.json() as Promise<AttendanceRecord[]>,
        classesRes.json() as Promise<ClassItem[]>,
        typesRes.json() as Promise<ClassType[]>,
      ])
      setAttendance(attendanceData)
      setClasses(classesData)
      setClassTypes(typesData)
    } catch {
      setError('Failed to load attendance history')
    } finally {
      setLoading(false)
    }
  }

  function getClassTypeName(classId: number): string {
    const cls = classes.find(c => c.id === classId)
    if (!cls) return 'Unknown'
    return classTypes.find(ct => ct.id === cls.class_type_id)?.name ?? 'Unknown'
  }

  function getClassTime(classId: number): string {
    const cls = classes.find(c => c.id === classId)
    if (!cls) return 'Unknown'
    return new Date(cls.start_time).toLocaleString()
  }

  if (loading) return <div>Loading...</div>
  if (error) return <div role="alert">{error}</div>

  return (
    <div>
      <h1>My Attendance</h1>
      <p>Total classes attended: {attendance.length}</p>

      {attendance.length === 0 ? (
        <p>No attendance records yet.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Class</th>
              <th>Class Date</th>
              <th>Checked In At</th>
            </tr>
          </thead>
          <tbody>
            {attendance.map(a => (
              <tr key={a.id}>
                <td data-label="Class">{getClassTypeName(a.class_id)}</td>
                <td data-label="Class Date">{getClassTime(a.class_id)}</td>
                <td data-label="Checked In">{new Date(a.checked_in_at).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
