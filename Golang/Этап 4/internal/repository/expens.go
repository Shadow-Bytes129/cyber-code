package repository

import (
 "database/sql"
 "time"

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
  "INSERT INTO expenses (amount, description, date, category_id) VALUES (?, ?, ?, ?)",
  e.Amount, e.Description, e.Date.Format("2006-01-02"), e.CategoryID,
 )
 if err != nil {
  return 0, err
 }
 return res.LastInsertId()
}

func (r *ExpenseRepository) GetAll(categoryID int64, fromDate, toDate string) ([]models.Expense, error) {
 query := `SELECT e.id, e.amount, e.description, e.date,
                     COALESCE(e.category_id, 0), COALESCE(c.name, '')
              FROM expenses e
              LEFT JOIN categories c ON c.id = e.category_id
              WHERE 1=1`
 var args []any

 if categoryID > 0 {
  query += " AND e.category_id = ?"
  args = append(args, categoryID)
 }
 if fromDate != "" {
  query += " AND e.date >= ?"
  args = append(args, fromDate)
 }
 if toDate != "" {
  query += " AND e.date <= ?"
  args = append(args, toDate)
 }

 query += " ORDER BY e.date DESC, e.id DESC"

 rows, err := r.db.Query(query, args...)
 if err != nil {
  return nil, err
 }
 defer rows.Close()

 var list []models.Expense
 for rows.Next() {
  var e models.Expense
  var dateStr string
  if err := rows.Scan(&e.ID, &e.Amount, &e.Description, &dateStr, &e.CategoryID, &e.CategoryName); err != nil {
   return nil, err
  }
  e.Date, _ = time.Parse("2006-01-02", dateStr)
  list = append(list, e)
 }
 return list, rows.Err()
}

func (r *ExpenseRepository) GetByID(id int64) (*models.Expense, error) {
 var e models.Expense
 var dateStr string
 err := r.db.QueryRow(
  `SELECT e.id, e.amount, e.description, e.date,
                COALESCE(e.category_id, 0), COALESCE(c.name, '')
         FROM expenses e
         LEFT JOIN categories c ON c.id = e.category_id
         WHERE e.id = ?`, id,
 ).Scan(&e.ID, &e.Amount, &e.Description, &dateStr, &e.CategoryID, &e.CategoryName)
 if err == sql.ErrNoRows {
  return nil, nil
 }
 if err != nil {
  return nil, err
 }
 e.Date, _ = time.Parse("2006-01-02", dateStr)
 return &e, nil
}

func (r *ExpenseRepository) Update(e *models.Expense) error {
 _, err := r.db.Exec(
  "UPDATE expenses SET amount = ?, description = ?, date = ?, category_id = ? WHERE id = ?",
  e.Amount, e.Description, e.Date.Format("2006-01-02"), e.CategoryID, e.ID,
 )
 return err
}

func (r *ExpenseRepository) Delete(id int64) error {
 _, err := r.db.Exec("DELETE FROM expenses WHERE id = ?", id)
 return err
}

func (r *ExpenseRepository) GetAllCategories() ([]models.Category, error) {
 rows, err := r.db.Query("SELECT id, name FROM categories ORDER BY id")
 if err != nil {
  return nil, err
 }
 defer rows.Close()
 var list []models.Category
 for rows.Next() {
  var c models.Category
  if err := rows.Scan(&c.ID, &c.Name); err != nil {
   return nil, err
  }
  list = append(list, c)
 }
 return list, rows.Err()
}

func (r *ExpenseRepository) GetTotal(categoryID int64, fromDate, toDate string) (float64, error) {
 query := "SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE 1=1"
 var args []any

 if categoryID > 0 {
  query += " AND category_id = ?"
  args = append(args, categoryID)
 }
 if fromDate != "" {
  query += " AND date >= ?"
  args = append(args, fromDate)
 }
 if toDate != "" {
  query += " AND date <= ?"
  args = append(args, toDate)
 }

 var total sql.NullFloat64
 err := r.db.QueryRow(query, args...).Scan(&total)
 if err != nil {
  return 0, err
 }
 return total.Float64, nil
}