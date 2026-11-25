package repository

import (
	"database/sql"
	"time"

	"expenses/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser — создание нового пользователя
func (r *UserRepository) CreateUser(email, passwordHash string) (int64, error) {
	res, err := r.db.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", email, passwordHash)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetUserByEmail — поиск пользователя для входа
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow("SELECT id, email, password_hash FROM users WHERE email = ?", email).
		Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// CreateSession — сохранение сессии в БД
func (r *UserRepository) CreateSession(sessionID string, userID int64, expiresAt time.Time) error {
	_, err := r.db.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, expiresAt)
	return err
}

// GetUserIDBySession — проверка сессии и срока ее действия
func (r *UserRepository) GetUserIDBySession(sessionID string) (int64, error) {
	var userID int64
	var expiresAt time.Time
	err := r.db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE id = ?", sessionID).Scan(&userID, &expiresAt)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	// Если сессия истекла — удаляем её и возвращаем 0
	if time.Now().After(expiresAt) {
		r.DeleteSession(sessionID)
		return 0, nil
	}
	return userID, nil
}

// DeleteSession — выход из системы (удаление сессии)
func (r *UserRepository) DeleteSession(sessionID string) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}
