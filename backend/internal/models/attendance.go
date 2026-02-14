package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Attendance struct {
	ID          int64
	ClassID     int64
	UserID      int64
	CheckedInAt time.Time
}

type AttendanceRepo struct {
	db *sql.DB
}

func NewAttendanceRepo(db *sql.DB) *AttendanceRepo {
	return &AttendanceRepo{db: db}
}

func (r *AttendanceRepo) RecordAttendance(a *Attendance) error {
	res, err := r.db.Exec(`INSERT INTO attendance (class_id, user_id) VALUES (?, ?)`,
		a.ClassID, a.UserID,
	)
	if err != nil {
		return fmt.Errorf("recording attendance: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	a.ID = id
	return nil
}

func (r *AttendanceRepo) ListByClass(classID int64) ([]Attendance, error) {
	rows, err := r.db.Query(`SELECT id, class_id, user_id, checked_in_at
		FROM attendance WHERE class_id = ? ORDER BY checked_in_at`, classID)
	if err != nil {
		return nil, fmt.Errorf("listing attendance by class: %w", err)
	}
	defer rows.Close()

	var records []Attendance
	for rows.Next() {
		var a Attendance
		if err := rows.Scan(&a.ID, &a.ClassID, &a.UserID, &a.CheckedInAt); err != nil {
			return nil, fmt.Errorf("scanning attendance: %w", err)
		}
		records = append(records, a)
	}
	return records, rows.Err()
}

func (r *AttendanceRepo) ListByUser(userID int64) ([]Attendance, error) {
	rows, err := r.db.Query(`SELECT id, class_id, user_id, checked_in_at
		FROM attendance WHERE user_id = ? ORDER BY checked_in_at`, userID)
	if err != nil {
		return nil, fmt.Errorf("listing attendance by user: %w", err)
	}
	defer rows.Close()

	var records []Attendance
	for rows.Next() {
		var a Attendance
		if err := rows.Scan(&a.ID, &a.ClassID, &a.UserID, &a.CheckedInAt); err != nil {
			return nil, fmt.Errorf("scanning attendance: %w", err)
		}
		records = append(records, a)
	}
	return records, rows.Err()
}
