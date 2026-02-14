package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Class struct {
	ID              int64
	ClassTypeID     int64
	InstructorID    int64
	StartTime       time.Time
	DurationMinutes int
	Capacity        int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ClassRepo struct {
	db *sql.DB
}

func NewClassRepo(db *sql.DB) *ClassRepo {
	return &ClassRepo{db: db}
}

func (r *ClassRepo) Create(c *Class) error {
	res, err := r.db.Exec(`INSERT INTO classes (class_type_id, instructor_id, start_time, duration_minutes, capacity)
		VALUES (?, ?, ?, ?, ?)`,
		c.ClassTypeID, c.InstructorID, c.StartTime, c.DurationMinutes, c.Capacity,
	)
	if err != nil {
		return fmt.Errorf("creating class: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	c.ID = id
	return nil
}

func (r *ClassRepo) GetByID(id int64) (*Class, error) {
	var c Class
	err := r.db.QueryRow(`SELECT id, class_type_id, instructor_id, start_time, duration_minutes, capacity, created_at, updated_at
		FROM classes WHERE id = ?`, id).
		Scan(&c.ID, &c.ClassTypeID, &c.InstructorID, &c.StartTime, &c.DurationMinutes, &c.Capacity, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ClassFilter holds optional filters for listing classes.
type ClassFilter struct {
	ClassTypeID  *int64
	InstructorID *int64
	From         *time.Time
	To           *time.Time
}

func (r *ClassRepo) List(filter ClassFilter) ([]Class, error) {
	query := `SELECT id, class_type_id, instructor_id, start_time, duration_minutes, capacity, created_at, updated_at
		FROM classes WHERE 1=1`
	var args []any

	if filter.ClassTypeID != nil {
		query += ` AND class_type_id = ?`
		args = append(args, *filter.ClassTypeID)
	}
	if filter.InstructorID != nil {
		query += ` AND instructor_id = ?`
		args = append(args, *filter.InstructorID)
	}
	if filter.From != nil {
		query += ` AND start_time >= ?`
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		query += ` AND start_time < ?`
		args = append(args, *filter.To)
	}

	query += ` ORDER BY start_time`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing classes: %w", err)
	}
	defer rows.Close()

	var classes []Class
	for rows.Next() {
		var c Class
		if err := rows.Scan(&c.ID, &c.ClassTypeID, &c.InstructorID, &c.StartTime, &c.DurationMinutes, &c.Capacity, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning class: %w", err)
		}
		classes = append(classes, c)
	}
	return classes, rows.Err()
}

func (r *ClassRepo) Update(c *Class) error {
	_, err := r.db.Exec(`UPDATE classes SET class_type_id=?, instructor_id=?, start_time=?, duration_minutes=?, capacity=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		c.ClassTypeID, c.InstructorID, c.StartTime, c.DurationMinutes, c.Capacity, c.ID,
	)
	if err != nil {
		return fmt.Errorf("updating class: %w", err)
	}
	return nil
}

func (r *ClassRepo) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM classes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting class: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
