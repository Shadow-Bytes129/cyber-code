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

// Create — добавление траты с привязкой к user_id
func (r *ExpenseRepository) Create(e *models.Expense) (int64, error) {
 res, err := r.db.Exec(
  "INSERT INTO expenses (user_id, amount, description, date, category_id) VALUES (?, ?, ?, ?, ?)",
  e.UserID, e.Amount, e.Description, e.Date.Format("2006-01-02"), e.CategoryID,
 )
 if err != nil {
  return 0, err
 }
 return res.LastInsertId()
}

// GetAll — получение трат только текущего пользователя с учетом фильтров
func (r *ExpenseRepository) GetAll(userID, categoryID int64, fromDate, toDate string) ([]models.Expense, error) {
 query := `SELECT e.id, e.amount, e.description, e.date,
                     COALESCE(e.category_id, 0), COALESCE(c.name, '')
              FROM expenses e
              LEFT JOIN categories c ON c.id = e.category_id
              WHERE e.user_id = ?`
 args := []any{userID}

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
  e.UserID = userID
  list = append(list, e)
 }
 return list, rows.Err()
}

// GetByID — получение траты по ID с проверкой, что она принадлежит пользователю
func (r *ExpenseRepository) GetByID(userID, id int64) (*models.Expense, error) {
 var e models.Expense
 var dateStr string
 err := r.db.QueryRow(
  `SELECT e.id, e.amount, e.description, e.date,
                COALESCE(e.category_id, 0), COALESCE(c.name, '')
         FROM expenses e
         LEFT JOIN categories c ON c.id = e.category_id
         WHERE e.id = ? AND e.user_id = ?`, id, userID,
 ).Scan(&e.ID, &e.Amount, &e.Description, &dateStr, &e.CategoryID, &e.CategoryName)
 if err == sql.ErrNoRows {
  return nil, nil
 }
 if err != nil {
  return nil, err
 }
 e.Date, _ = time.Parse("2006-01-02", dateStr)
 e.UserID = userID
 return &e, nil
}

// Update — обновление только своей траты
func (r *ExpenseRepository) Update(e *models.Expense) error {
 _, err := r.db.Exec(
  "UPDATE expenses SET amount = ?, description = ?, date = ?, category_id = ? WHERE id = ? AND user_id = ?",
  e.Amount, e.Description, e.Date.Format("2006-01-02"), e.CategoryID, e.ID, e.UserID,
 )
 return err
}

// Delete — удаление только своей траты
func (r *ExpenseRepository) Delete(userID, id int64) error {
 _, err := r.db.Exec("DELETE FROM expenses WHERE id = ? AND user_id = ?", id, userID)
 return err
}

// GetAllCategories — категории общие для всех пользователей
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

// GetTotal — подсчет суммы только для текущего пользователя
func (r *ExpenseRepository) GetTotal(userID, categoryID int64, fromDate, toDate string) (float64, error) {
 query := "SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE user_id = ?"
 args := []any{userID}

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