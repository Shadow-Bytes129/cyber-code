package main

import (
	"fmt"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Добро пожаловать на главную страницу!")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Это простой HTTP-сервер на Go. Проект демонстрирует работу с маршрутами.")
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "pong")
}

func main() {
	host := "localhost"
	port := "8080"

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/ping", pingHandler)

	baseURL := fmt.Sprintf("http://%s:%s", host, port)
	fmt.Println("Сервер запущен!")
	fmt.Printf("Главная: %s\n", baseURL)
	fmt.Printf("About: %s/about\n", baseURL)
	fmt.Printf("Ping: %s/ping\n", baseURL)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Ошибка при запуске сервера: %v\n", err)
	}
}
