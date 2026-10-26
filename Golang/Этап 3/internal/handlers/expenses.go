package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"expenses/internal/models"
	"expenses/internal/repository"
)

type ExpenseHandler struct {
	repo *repository.ExpenseRepository
	tmpl *template.Template
}

func NewExpenseHandler(repo *repository.ExpenseRepository) *ExpenseHandler {
	return &ExpenseHandler{
		repo: repo,
		tmpl: template.Must(template.ParseGlob("web/templates/*.html")),
	}
}

func (h *ExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
	expenses, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.tmpl.ExecuteTemplate(w, "index.html", map[string]any{
		"Expenses": expenses,
	})
}

func (h *ExpenseHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	h.tmpl.ExecuteTemplate(w, "form.html", map[string]any{
		"Expense": models.Expense{},
		"IsEdit":  false,
	})
}

func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	if _, err := h.repo.Create(&models.Expense{
		Amount:      amount,
		Description: r.FormValue("description"),
		Date:        r.FormValue("date"),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *ExpenseHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	e, err := h.repo.GetByID(id)
	if err != nil || e == nil {
		http.NotFound(w, r)
		return
	}
	h.tmpl.ExecuteTemplate(w, "form.html", map[string]any{
		"Expense": e,
		"IsEdit":  true,
	})
}

func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	r.ParseForm()
	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err := h.repo.Update(&models.Expense{
		ID:          id,
		Amount:      amount,
		Description: r.FormValue("description"),
		Date:        r.FormValue("date"),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.repo.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}
