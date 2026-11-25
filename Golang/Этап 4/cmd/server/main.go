package main

import (
    "database/sql"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "os"

    _ "modernc.org/sqlite"

    "expenses/internal/handlers"
    "expenses/internal/repository"
)

func main() {
    // 1. Открываем БД
    db, err := sql.Open("sqlite", "expenses.db?_pragma=foreign_keys(1)")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    db.SetMaxOpenConns(1) // SQLite: избегаем блокировок

    // 2. Миграции
    if err := applyMigrations(db); err != nil {
        log.Fatal("migration: ", err)
    }

    // 3. Слои
    tmpl := template.Must(template.ParseGlob("web/templates/*.html"))
    repo := repository.NewExpenseRepository(db)
    h := handlers.NewExpenseHandler(repo, tmpl)

    // 4. Маршруты (Go 1.22+ pattern syntax)
    mux := http.NewServeMux()
    mux.HandleFunc("GET /{$}", h.List)
    mux.HandleFunc("GET /expenses", h.List)
    mux.HandleFunc("GET /expenses/new", h.NewForm)
    mux.HandleFunc("POST /expenses", h.Create)
    mux.HandleFunc("GET /expenses/{id}/edit", h.EditForm)
    mux.HandleFunc("POST /expenses/{id}", h.Update)
    mux.HandleFunc("POST /expenses/{id}/delete", h.Delete)

    mux.Handle("GET /static/",
        http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

    fmt.Println("===================================================")
    fmt.Println("Сервер запущен!")
    fmt.Println("Список трат:    http://localhost:8080/expenses")
    fmt.Println("Добавить трату: http://localhost:8080/expenses/new")
    fmt.Println("===================================================")

    log.Fatal(http.ListenAndServe(":8080", mux))
}

func applyMigrations(db *sql.DB) error {
    files := []string{"migrations/001_init.sql"}
    for _, f := range files {
        b, err := os.ReadFile(f)
        if err != nil {
            return fmt.Errorf("read %s: %w", f, err)
        }
        if _, err := db.Exec(string(b)); err != nil {
            return fmt.Errorf("exec %s: %w", f, err)
        }
    }
    return nil
}