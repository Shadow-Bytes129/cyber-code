package handlers

import (
 "encoding/json"
 "html/template"
 "net/http"

 "expenses/internal/repository"
)

type StatsHandler struct {
 repo *repository.ExpenseRepository
 tmpl *template.Template
}

func NewStatsHandler(repo *repository.ExpenseRepository, tmpl *template.Template) *StatsHandler {
 return &StatsHandler{repo: repo, tmpl: tmpl}
}

// Page — отрисовка HTML-страницы со статистикой
func (h *StatsHandler) Page(w http.ResponseWriter, r *http.Request) {
 h.tmpl.ExecuteTemplate(w, "layout", map[string]any{
  "Page": "stats",
 })
}

// APIByCategory — JSON-эндпоинт для круговой диаграммы
func (h *StatsHandler) APIByCategory(w http.ResponseWriter, r *http.Request) {
 userID := getUserID(r)
 year := r.URL.Query().Get("year")
 stats, err := h.repo.GetStatsByCategory(userID, year)
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.Header().Set("Content-Type", "application/json")
 json.NewEncoder(w).Encode(stats)
}

// APIByMonth — JSON-эндпоинт для столбчатой диаграммы
func (h *StatsHandler) APIByMonth(w http.ResponseWriter, r *http.Request) {
 userID := getUserID(r)
 year := r.URL.Query().Get("year")
 stats, err := h.repo.GetStatsByMonth(userID, year)
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.Header().Set("Content-Type", "application/json")
 json.NewEncoder(w).Encode(stats)
}