package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type ctxKey string

const ctxUserIDKey ctxKey = "userID"
const ctxEmailKey ctxKey = "email"

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (app *App) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "missing Authorization header")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			writeError(w, http.StatusUnauthorized, "expected format: Authorization: Bearer <token>")
			return
		}

		claims, err := ParseToken(app.jwtSecret, parts[1])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserIDKey, claims.Sub)
		ctx = context.WithValue(ctx, ctxEmailKey, claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func userIDFromContext(r *http.Request) string {
	if v, ok := r.Context().Value(ctxUserIDKey).(string); ok {
		return v
	}
	return ""
}
