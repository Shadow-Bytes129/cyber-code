package models

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

type Session struct {
	ID        string
	UserID    int64
	ExpiresAt time.Time
}

type Expense struct {
	ID           int64
	UserID       int64 // Привязка к пользователю
	Amount       float64
	Description  string
	Date         time.Time
	CategoryID   int64
	CategoryName string
}

type Category struct {
	ID   int64
	Name string
}
