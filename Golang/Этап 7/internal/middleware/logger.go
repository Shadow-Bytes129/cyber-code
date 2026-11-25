package middleware

import (
 "log"
 "net/http"
 "time"
)

// responseWriter — обертка для http.ResponseWriter, чтобы перехватить статус-код ответа
type responseWriter struct {
 http.ResponseWriter
 statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
 rw.statusCode = code
 rw.ResponseWriter.WriteHeader(code)
}

// Logger — логирует каждый запрос и восстанавливается после паники
func Logger(next http.Handler) http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  start := time.Now()
  rw := &responseWriter{w, http.StatusOK} // По умолчанию 200

  // 1. Защита от паники (recover)
  defer func() {
   if err := recover(); err != nil {
    log.Printf("PANIC: %v", err)
    http.Error(rw, "Внутренняя ошибка сервера", http.StatusInternalServerError)
   }
   // 2. Логирование запроса
   duration := time.Since(start)
   log.Printf("%s %s %d %v", r.Method, r.URL.Path, rw.statusCode, duration)
  }()

  // Передаем запрос дальше
  next.ServeHTTP(rw, r)
 })
}