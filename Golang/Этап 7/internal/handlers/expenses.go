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
  h.renderError(w, http.StatusInternalServerError, "Ошибка получения списка трат")
  return
 }
 categories, err := h.repo.GetAllCategories()
 if err != nil {
  h.renderError(w, http.StatusInternalServerError, "Ошибка загрузки категорий")
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
  h.renderError(w, http.StatusNotFound, "Трата не найдена")
  return
 }
 e, err := h.repo.GetByID(userID, id)
 if err != nil || e == nil {
  h.renderError(w, http.StatusNotFound, "Трата не найдена")
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
 e, errMsg := h.parseForm(r)
 if errMsg != "" {
  // Ошибка валидации: возвращаем форму с введенными данными
  categories, _ := h.repo.GetAllCategories()
  e.ErrorMsg = errMsg
  h.render(w, "layout", map[string]any{
   "Page":       "form",
   "IsEdit":     false,
   "Expense":    e,
   "Categories": categories,
  })
  return
 }
 e.UserID = userID
 if _, err := h.repo.Create(e); err != nil {
  h.renderError(w, http.StatusInternalServerError, "Ошибка сохранения траты")
  return
 }
 http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
 userID := getUserID(r)
 id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
 if err != nil {
  h.renderError(w, http.StatusNotFound, "Трата не найдена")
  return
 }
 e, errMsg := h.parseForm(r)
 e.ID = id
 e.UserID = userID
 
 if errMsg != "" {
  categories, _ := h.repo.GetAllCategories()
  e.ErrorMsg = errMsg
  h.render(w, "layout", map[string]any{
   "Page":       "form",
   "IsEdit":     true,
   "Expense":    e,
   "Categories": categories,
  })
  return
 }
 
 if err := h.repo.Update(e); err != nil {
  h.renderError(w, http.StatusInternalServerError, "Ошибка обновления траты")
  return
 }
 http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
 userID := getUserID(r)
 id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
 if err != nil {
  h.renderError(w, http.StatusNotFound, "Трата не найдена")
  return
 }
 if err := h.repo.Delete(userID, id); err != nil {
  h.renderError(w, http.StatusInternalServerError, "Ошибка удаления траты")
  return
 }
 http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// parseForm — теперь возвращает текст ошибки вместо стандартного error
func (h *ExpenseHandler) parseForm(r *http.Request) (*models.Expense, string) {
 if err := r.ParseForm(); err != nil {
  return &models.Expense{}, "Ошибка отправки формы"
 }
 
 amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
 if err != nil || amount <= 0 {
  return &models.Expense{Description: r.FormValue("description")}, "Сумма должна быть положительным числом"
 }
 
 date, err := time.Parse("2006-01-02", r.FormValue("date"))
 if err != nil {
  return &models.Expense{Amount: amount, Description: r.FormValue("description")}, "Введите корректную дату"
 }
 
 desc := r.FormValue("description")
 if desc == "" {
  return &models.Expense{Amount: amount, Date: date}, "Описание не может быть пустым"
 }
 
 catID, _ := strconv.ParseInt(r.FormValue("category_id"), 10, 64)
 return &models.Expense{
  Amount:      amount,
  Description: desc,
  Date:        date,
  CategoryID:  catID,
 }, ""
}

// render — вспомогательный метод для отрисовки шаблонов
func (h *ExpenseHandler) render(w http.ResponseWriter, name string, data any) {
 if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
  fmt.Println("Template error:", err)
  http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
 }
}

// renderError — рендерит кастомную страницу ошибки
func (h *ExpenseHandler) renderError(w http.ResponseWriter, statusCode int, message string) {
 w.WriteHeader(statusCode)
 h.tmpl.ExecuteTemplate(w, "layout", map[string]any{
  "Page":       "error",
  "ErrorCode":  statusCode,
  "ErrorMessage": message,
 })
}