package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"

	"expenses/internal/handlers"
	"expenses/internal/repository"
)

func main() {
	// 1. Открыть БД
	db, err := sql.Open("sqlite", "expenses.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1) // для SQLite с файлом — избегаем блокировок

	// 2. Применить миграцию при старте
	migration, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		log.Fatal("migration read: ", err)
	}
	if _, err := db.Exec(string(migration)); err != nil {
		log.Fatal("migration exec: ", err)
	}

	// 3. Собрать слои
	repo := repository.NewExpenseRepository(db)
	h := handlers.NewExpenseHandler(repo)

	// 4. Маршруты
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

	// Красивый вывод ссылок в терминал
	fmt.Println("======================================================")
	fmt.Println("🚀 Сервер запущен! Доступные страницы (GET):")
	fmt.Println("👉 Главная / Список трат:  http://localhost:8080/")
	fmt.Println("👉 Список трат:            http://localhost:8080/expenses")
	fmt.Println("👉 Добавить новую трату:   http://localhost:8080/expenses/new")
	fmt.Println("======================================================")

	// Запуск сервера
	log.Fatal(http.ListenAndServe(":8080", mux))
}
