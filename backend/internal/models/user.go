package models

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	ID               int64
	Name             string
	Email            string
	Phone            string
	Role             string
	PasswordHash     string
	MembershipType   *string
	MembershipStatus *string
	EmergencyContact *string
	JoinDate         *string
	ExpectedBalance  float64
	DeletedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(u *User) error {
	res, err := r.db.Exec(`INSERT INTO users (name, email, phone, role, password_hash, membership_type, membership_status, emergency_contact, join_date, expected_balance)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.Name, u.Email, u.Phone, u.Role, u.PasswordHash,
		u.MembershipType, u.MembershipStatus, u.EmergencyContact, u.JoinDate, u.ExpectedBalance,
	)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	u.ID = id
	return nil
}

func (r *UserRepo) GetByID(id int64) (*User, error) {
	return r.scanUser(r.db.QueryRow(`SELECT id, name, email, phone, role, password_hash, membership_type, membership_status, emergency_contact, join_date, expected_balance, deleted_at, created_at, updated_at
		FROM users WHERE id = ?`, id))
}

func (r *UserRepo) GetByEmail(email string) (*User, error) {
	return r.scanUser(r.db.QueryRow(`SELECT id, name, email, phone, role, password_hash, membership_type, membership_status, emergency_contact, join_date, expected_balance, deleted_at, created_at, updated_at
		FROM users WHERE email = ?`, email))
}

func (r *UserRepo) List() ([]User, error) {
	rows, err := r.db.Query(`SELECT id, name, email, phone, role, password_hash, membership_type, membership_status, emergency_contact, join_date, expected_balance, deleted_at, created_at, updated_at
		FROM users WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := r.scanUserRow(rows, &u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepo) Update(u *User) error {
	_, err := r.db.Exec(`UPDATE users SET name=?, email=?, phone=?, role=?, password_hash=?, membership_type=?, membership_status=?, emergency_contact=?, join_date=?, expected_balance=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		u.Name, u.Email, u.Phone, u.Role, u.PasswordHash,
		u.MembershipType, u.MembershipStatus, u.EmergencyContact, u.JoinDate, u.ExpectedBalance,
		u.ID,
	)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	return nil
}

func (r *UserRepo) SoftDelete(id int64) error {
	res, err := r.db.Exec(`UPDATE users SET deleted_at=CURRENT_TIMESTAMP, updated_at=CURRENT_TIMESTAMP WHERE id=? AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("soft deleting user: %w", err)
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

func (r *UserRepo) scanUser(row *sql.Row) (*User, error) {
	var u User
	err := row.Scan(
		&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role, &u.PasswordHash,
		&u.MembershipType, &u.MembershipStatus, &u.EmergencyContact, &u.JoinDate,
		&u.ExpectedBalance, &u.DeletedAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func (r *UserRepo) scanUserRow(s scannable, u *User) error {
	return s.Scan(
		&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role, &u.PasswordHash,
		&u.MembershipType, &u.MembershipStatus, &u.EmergencyContact, &u.JoinDate,
		&u.ExpectedBalance, &u.DeletedAt, &u.CreatedAt, &u.UpdatedAt,
	)
}
