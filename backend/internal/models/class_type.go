package models

import (
	"database/sql"
	"fmt"
	"time"
)

type ClassType struct {
	ID          int64
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ClassTypeRepo struct {
	db *sql.DB
}

func NewClassTypeRepo(db *sql.DB) *ClassTypeRepo {
	return &ClassTypeRepo{db: db}
}

func (r *ClassTypeRepo) Create(ct *ClassType) error {
	res, err := r.db.Exec(`INSERT INTO class_types (name, description) VALUES (?, ?)`,
		ct.Name, ct.Description,
	)
	if err != nil {
		return fmt.Errorf("creating class type: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	ct.ID = id
	return nil
}

func (r *ClassTypeRepo) GetByID(id int64) (*ClassType, error) {
	var ct ClassType
	err := r.db.QueryRow(`SELECT id, name, description, created_at, updated_at FROM class_types WHERE id = ?`, id).
		Scan(&ct.ID, &ct.Name, &ct.Description, &ct.CreatedAt, &ct.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

func (r *ClassTypeRepo) List() ([]ClassType, error) {
	rows, err := r.db.Query(`SELECT id, name, description, created_at, updated_at FROM class_types ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("listing class types: %w", err)
	}
	defer rows.Close()

	var types []ClassType
	for rows.Next() {
		var ct ClassType
		if err := rows.Scan(&ct.ID, &ct.Name, &ct.Description, &ct.CreatedAt, &ct.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning class type: %w", err)
		}
		types = append(types, ct)
	}
	return types, rows.Err()
}

func (r *ClassTypeRepo) Update(ct *ClassType) error {
	_, err := r.db.Exec(`UPDATE class_types SET name=?, description=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		ct.Name, ct.Description, ct.ID,
	)
	if err != nil {
		return fmt.Errorf("updating class type: %w", err)
	}
	return nil
}

func (r *ClassTypeRepo) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM class_types WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting class type: %w", err)
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
