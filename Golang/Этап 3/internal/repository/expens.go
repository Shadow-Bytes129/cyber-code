package repository

import (
	"database/sql"

	"expenses/internal/models"
)

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) Create(e *models.Expense) (int64, error) {
	res, err := r.db.Exec(
		"INSERT INTO expenses (amount, description, date) VALUES (?, ?, ?)",
		e.Amount, e.Description, e.Date,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ExpenseRepository) GetAll() ([]models.Expense, error) {
	rows, err := r.db.Query(
		"SELECT id, amount, description, date FROM expenses ORDER BY date DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // ОБЯЗАТЕЛЬНО

	var list []models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.ID, &e.Amount, &e.Description, &e.Date); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *ExpenseRepository) GetByID(id int64) (*models.Expense, error) {
	var e models.Expense
	err := r.db.QueryRow(
		"SELECT id, amount, description, date FROM expenses WHERE id = ?", id,
	).Scan(&e.ID, &e.Amount, &e.Description, &e.Date)
	if err == sql.ErrNoRows {
		return nil, nil // не найдено — не ошибка
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ExpenseRepository) Update(e *models.Expense) error {
	_, err := r.db.Exec(
		"UPDATE expenses SET amount = ?, description = ?, date = ? WHERE id = ?",
		e.Amount, e.Description, e.Date, e.ID,
	)
	return err
}

func (r *ExpenseRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM expenses WHERE id = ?", id)
	return err
}
