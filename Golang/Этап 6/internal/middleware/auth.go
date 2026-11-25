package middleware

import (
 "context"
 "net/http"

 "expenses/internal/repository"
)

// contextKey — свой тип для ключей контекста, чтобы избежать коллизий
type contextKey string

// UserIDKey — ключ, по которому мы будем доставать ID пользователя из контекста
const UserIDKey contextKey = "userID"

// RequireAuth — проверяет наличие сессии перед доступом к страницам
func RequireAuth(userRepo *repository.UserRepository, next http.HandlerFunc) http.HandlerFunc {
 return func(w http.ResponseWriter, r *http.Request) {
  // 1. Пытаемся получить куки с ID сессии
  cookie, err := r.Cookie("session_id")
  if err != nil {
   // Если куки нет — отправляем на страницу входа
   http.Redirect(w, r, "/login", http.StatusSeeOther)
   return
  }

  // 2. Проверяем сессию в базе данных
  userID, err := userRepo.GetUserIDBySession(cookie.Value)
  if err != nil || userID == 0 {
   // Если сессия невалидна или истекла — тоже на страницу входа
   http.Redirect(w, r, "/login", http.StatusSeeOther)
   return
  }

  // 3. Кладем ID пользователя в контекст запроса
  ctx := context.WithValue(r.Context(), UserIDKey, userID)
  
  // 4. Передаем запрос дальше (в наш обработчик), но уже с обновленным контекстом
  next(w, r.WithContext(ctx))
 }
}