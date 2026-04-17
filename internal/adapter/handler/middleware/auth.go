package middleware

import (
	"context"
	"net/http"

	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/auth"
	"github.com/aikowocki/yandex-go-first-diploma/internal/pkg/response"
)

type contextKey string

const userIDKey contextKey = "user_id"

func WithAuth(jwtManager *auth.JWTManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
			userID, err := jwtManager.ParseUserID(cookie.Value)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}

func MustGetUserID(ctx context.Context) int64 {
	userID, ok := GetUserID(ctx)
	if !ok {
		// По сути ошибка сервиса если юзкейс требует авторизации, но к этому шагу ее не было
		panic("userID not found in context: auth middleware missing")
	}
	return userID
}
