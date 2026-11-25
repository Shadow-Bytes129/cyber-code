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
 UserID       int64
 Amount       float64
 Description  string
 Date         time.Time
 CategoryID   int64
 CategoryName string
 ErrorMsg     string // Новое поле для ошибок валидации
}

type Category struct {
 ID   int64
 Name string
}

type CategoryStat struct {
 Name  string  `json:"name"`
 Total float64 `json:"total"`
}

type MonthStat struct {
 Month string  `json:"month"`
 Total float64 `json:"total"`
}