package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	_ "modernc.org/sqlite"

	"expenses/internal/handlers"
	"expenses/internal/middleware"
	"expenses/internal/repository"
)

func main() {
	db, err := sql.Open("sqlite", "expenses.db?_pragma=foreign_keys(1)")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if err := applyMigrations(db); err != nil {
		log.Fatal("migration: ", err)
	}

	tmpl := template.Must(template.ParseGlob("web/templates/*.html"))

	expenseRepo := repository.NewExpenseRepository(db)
	userRepo := repository.NewUserRepository(db)

	h := handlers.NewExpenseHandler(expenseRepo, tmpl)
	auth := handlers.NewAuthHandler(userRepo, tmpl)
	stats := handlers.NewStatsHandler(expenseRepo, tmpl)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /register", auth.RegisterForm)
	mux.HandleFunc("POST /register", auth.Register)
	mux.HandleFunc("GET /login", auth.LoginForm)
	mux.HandleFunc("POST /login", auth.Login)
	mux.HandleFunc("POST /logout", auth.Logout)

	mux.HandleFunc("GET /{$}", middleware.RequireAuth(userRepo, h.List))
	mux.HandleFunc("GET /expenses", middleware.RequireAuth(userRepo, h.List))
	mux.HandleFunc("GET /expenses/new", middleware.RequireAuth(userRepo, h.NewForm))
	mux.HandleFunc("POST /expenses", middleware.RequireAuth(userRepo, h.Create))
	mux.HandleFunc("GET /expenses/{id}/edit", middleware.RequireAuth(userRepo, h.EditForm))
	mux.HandleFunc("POST /expenses/{id}", middleware.RequireAuth(userRepo, h.Update))
	mux.HandleFunc("POST /expenses/{id}/delete", middleware.RequireAuth(userRepo, h.Delete))

	mux.HandleFunc("GET /stats", middleware.RequireAuth(userRepo, stats.Page))
	mux.HandleFunc("GET /api/stats/by-category", middleware.RequireAuth(userRepo, stats.APIByCategory))
	mux.HandleFunc("GET /api/stats/by-month", middleware.RequireAuth(userRepo, stats.APIByMonth))

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Оборачиваем ВЕСЬ роутер в middleware.Logger
	loggedMux := middleware.Logger(mux)

	fmt.Println("===================================================")
	fmt.Println("Сервер запущен!")
	fmt.Println("Вход:           http://localhost:8080/login")
	fmt.Println("Регистрация:    http://localhost:8080/register")
	fmt.Println("===================================================")

	log.Fatal(http.ListenAndServe(":8080", loggedMux))
}

func applyMigrations(db *sql.DB) error {
	files := []string{
		"migrations/001_init.sql",
		"migrations/002_categories.sql",
		"migrations/003_users.sql",
	}
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}
		_, err = db.Exec(string(b))
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("exec %s: %w", file, err)
		}
	}
	return nil
}
