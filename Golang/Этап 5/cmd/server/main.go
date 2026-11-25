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
	// 1. Открываем БД
	db, err := sql.Open("sqlite", "expenses.db?_pragma=foreign_keys(1)")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	// 2. Миграции
	if err := applyMigrations(db); err != nil {
		log.Fatal("migration: ", err)
	}

	// 3. Слои
	tmpl := template.Must(template.ParseGlob("web/templates/*.html"))

	// Репозитории
	expenseRepo := repository.NewExpenseRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Обработчики
	h := handlers.NewExpenseHandler(expenseRepo, tmpl)
	auth := handlers.NewAuthHandler(userRepo, tmpl)

	// 4. Маршруты
	mux := http.NewServeMux()

	// Публичные маршруты 
	mux.HandleFunc("GET /register", auth.RegisterForm)
	mux.HandleFunc("POST /register", auth.Register)
	mux.HandleFunc("GET /login", auth.LoginForm)
	mux.HandleFunc("POST /login", auth.Login)
	mux.HandleFunc("POST /logout", auth.Logout)

	// Защищенные маршруты 
	mux.HandleFunc("GET /{$}", middleware.RequireAuth(userRepo, h.List))
	mux.HandleFunc("GET /expenses", middleware.RequireAuth(userRepo, h.List))
	mux.HandleFunc("GET /expenses/new", middleware.RequireAuth(userRepo, h.NewForm))
	mux.HandleFunc("POST /expenses", middleware.RequireAuth(userRepo, h.Create))
	mux.HandleFunc("GET /expenses/{id}/edit", middleware.RequireAuth(userRepo, h.EditForm))
	mux.HandleFunc("POST /expenses/{id}", middleware.RequireAuth(userRepo, h.Update))
	mux.HandleFunc("POST /expenses/{id}/delete", middleware.RequireAuth(userRepo, h.Delete))

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	fmt.Println("===================================================")
	fmt.Println("Сервер запущен!")
	fmt.Println("Вход:           http://localhost:8080/login")
	fmt.Println("Регистрация:    http://localhost:8080/register")
	fmt.Println("===================================================")

	log.Fatal(http.ListenAndServe(":8080", mux))
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
		if err != nil {
			if strings.Contains(err.Error(), "duplicate column name") {
				continue
			}
			return fmt.Errorf("exec %s: %w", file, err)
		}
	}
	return nil
}
