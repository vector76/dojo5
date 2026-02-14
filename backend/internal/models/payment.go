package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Payment struct {
	ID         int64
	UserID     int64
	Amount     float64
	Date       string
	Note       *string
	RecordedBy int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PaymentRepo struct {
	db *sql.DB
}

func NewPaymentRepo(db *sql.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) RecordPayment(p *Payment) error {
	res, err := r.db.Exec(`INSERT INTO payments (user_id, amount, date, note, recorded_by) VALUES (?, ?, ?, ?, ?)`,
		p.UserID, p.Amount, p.Date, p.Note, p.RecordedBy,
	)
	if err != nil {
		return fmt.Errorf("recording payment: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	p.ID = id
	return nil
}

func (r *PaymentRepo) ListByUser(userID int64) ([]Payment, error) {
	rows, err := r.db.Query(`SELECT id, user_id, amount, date, note, recorded_by, created_at, updated_at
		FROM payments WHERE user_id = ? ORDER BY date DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("listing payments by user: %w", err)
	}
	defer rows.Close()

	var payments []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.UserID, &p.Amount, &p.Date, &p.Note, &p.RecordedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning payment: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

// GetBalance returns the outstanding balance for a user.
// Balance = user.expected_balance - SUM(payments.amount).
func (r *PaymentRepo) GetBalance(userID int64) (float64, error) {
	var expectedBalance float64
	var totalPaid sql.NullFloat64

	err := r.db.QueryRow(`SELECT u.expected_balance, SUM(p.amount)
		FROM users u
		LEFT JOIN payments p ON p.user_id = u.id
		WHERE u.id = ?
		GROUP BY u.id`, userID).
		Scan(&expectedBalance, &totalPaid)
	if err != nil {
		return 0, err
	}

	if totalPaid.Valid {
		return expectedBalance - totalPaid.Float64, nil
	}
	return expectedBalance, nil
}
