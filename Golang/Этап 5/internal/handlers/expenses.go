package handlers

import (
 "fmt"
 "html/template"
 "net/http"
 "strconv"
 "time"

 "expenses/internal/middleware"
 "expenses/internal/models"
 "expenses/internal/repository"
)

type ExpenseHandler struct {
 repo *repository.ExpenseRepository
 tmpl *template.Template
}

func NewExpenseHandler(repo *repository.ExpenseRepository, tmpl *template.Template) *ExpenseHandler {
 return &ExpenseHandler{repo: repo, tmpl: tmpl}
}

// getUserID — вспомогательная функция для получения ID пользователя из контекста
func getUserID(r *http.Request) int64 {
 id, ok := r.Context().Value(middleware.UserIDKey).(int64)
 if !ok {
  return 0
 }
 return id
}

func (h *ExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
 userID := getUserID(r)
 filterID, _ := strconv.ParseInt(r.URL.Query().Get("category"), 10, 64)
 fromDate := r.URL.Query().Get("from")
 toDate := r.URL.Query().Get("to")

 expenses, err := h.repo.GetAll(userID, filterID, fromDate, toDate)
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }
 categories, err := h.repo.GetAllCategories()
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }
 total, _ := h.repo.GetTotal(userID, filterID, fromDate, toDate)

 h.render(w, "layout", map[string]any{
  "Page":       "list",
  "Expenses":   expenses,
  "Categories": categories,
  "FilterID":   filterID,
  "FromDate":   fromDate,
  "ToDate":     toDate,
  "Total":      total,
 })
}

func (h *ExpenseHandler) NewForm(w http.ResponseWriter, r *http.Request) {
 categories, _ := h.repo.GetAllCategories()
 h.render(w, "layout", map[string]any{
  "Page":       "form",
  "IsEdit":     false,
  "Expense":    models.Expense{Date: time.Now()},
  "Categories": categories,
 })
}

func (h *ExpenseHandler) EditForm(w http.ResponseWriter, r *http.Request) {
 userID := getUserID(r)
 id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
 if err != nil {
  http.NotFound(w, r)
  return
 }
 e, err := h.repo.GetByID(userID, id)
 if err != nil || e == nil {
  http.NotFound(w, r)
  return
 }
 categories, _ := h.repo.GetAllCategories()
 h.render(w, "layout", map[string]any{
  "Page":       "form",
  "IsEdit":     true,
  "Expense":    e,
  "Categories": categories,
 })
}

func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
 userID := getUserID(r)
 e, err := h.parseForm(r)
 if err != nil {
  http.Error(w, err.Error(), http.StatusBadRequest)
  return
 }
 e.UserID = userID
 if _, err := h.repo.Create(e); err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }
 http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
 userID := getUserID(r)
 id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
 if err != nil {
  http.NotFound(w, r)
  return
 }
 e, err := h.parseForm(r)
 if err != nil {
  http.Error(w, err.Error(), http.StatusBadRequest)
  return
 }
 e.ID = id
 e.UserID = userID
 if err := h.repo.Update(e); err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }
 http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
 userID := getUserID(r)
 id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
 if err != nil {
  http.NotFound(w, r)
  return
 }
 if err := h.repo.Delete(userID, id); err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }
 http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *ExpenseHandler) parseForm(r *http.Request) (*models.Expense, error) {
 if err := r.ParseForm(); err != nil {
  return nil, fmt.Errorf("bad form: %w", err)
 }
 amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
 if err != nil || amount <= 0 {
  return nil, fmt.Errorf("invalid amount")
 }
 date, err := time.Parse("2006-01-02", r.FormValue("date"))
 if err != nil {
  return nil, fmt.Errorf("invalid date")
 }
 desc := r.FormValue("description")
 if desc == "" {
  return nil, fmt.Errorf("invalid description")
 }
 catID, _ := strconv.ParseInt(r.FormValue("category_id"), 10, 64)
 return &models.Expense{
  Amount:      amount,
  Description: desc,
  Date:        date,
  CategoryID:  catID,
 }, nil
}

func (h *ExpenseHandler) render(w http.ResponseWriter, name string, data any) {
 if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
  http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
 }
}