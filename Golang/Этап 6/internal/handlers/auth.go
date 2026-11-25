package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"net/http"
	"time"

	"expenses/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
	tmpl     *template.Template
}

func NewAuthHandler(repo *repository.UserRepository, tmpl *template.Template) *AuthHandler {
	return &AuthHandler{userRepo: repo, tmpl: tmpl}
}

// RegisterForm — показ формы регистрации
func (h *AuthHandler) RegisterForm(w http.ResponseWriter, r *http.Request) {
	h.tmpl.ExecuteTemplate(w, "layout", map[string]any{"Page": "register"})
}

// Register — создание нового пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Хэшируем пароль (cost = 10 по умолчанию)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка при хэшировании", http.StatusInternalServerError)
		return
	}

	_, err = h.userRepo.CreateUser(email, string(hash))
	if err != nil {
		// Если email уже занят (ошибка UNIQUE constraint в БД)
		http.Error(w, "Пользователь с таким email уже существует", http.StatusBadRequest)
		return
	}

	// После успешной регистрации отправляем на страницу входа
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// LoginForm — показ формы входа
func (h *AuthHandler) LoginForm(w http.ResponseWriter, r *http.Request) {
	h.tmpl.ExecuteTemplate(w, "layout", map[string]any{"Page": "login"})
}

// Login — проверка пароля и создание сессии
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка формы", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.userRepo.GetUserByEmail(email)
	if err != nil || user == nil {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	// Сравниваем введенный пароль с хэшем из БД
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	// Генерируем случайный ID сессии (32 байта)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		http.Error(w, "Ошибка генерации сессии", http.StatusInternalServerError)
		return
	}
	sessionID := hex.EncodeToString(b)

	expiresAt := time.Now().Add(24 * time.Hour) // Сессия живет 24 часа
	if err := h.userRepo.CreateSession(sessionID, user.ID, expiresAt); err != nil {
		http.Error(w, "Ошибка создания сессии", http.StatusInternalServerError)
		return
	}

	// Устанавливаем Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  expiresAt,
		HttpOnly: true, // Защита от XSS (JavaScript не сможет прочитать куку)
		Path:     "/",
	})

	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// Logout — выход из системы
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		// Удаляем сессию из БД
		h.userRepo.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
