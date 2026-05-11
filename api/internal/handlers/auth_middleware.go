package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/sunkencosts/mirror-me/internal/jwtauth"
)

type contextKey string

const claimsKey contextKey = "claims"

func RequireAuth(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := extractClaims(r, jwtSecret)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromContext(ctx context.Context) (jwtauth.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(jwtauth.Claims)
	return claims, ok
}

const bearerPrefix = "Bearer "

func extractClaims(r *http.Request, secret []byte) (jwtauth.Claims, bool) {
	var tokenStr string
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, bearerPrefix) {
		tokenStr = strings.TrimPrefix(auth, bearerPrefix)
	} else if c, err := r.Cookie("auth_token"); err == nil {
		tokenStr = c.Value
	}
	if tokenStr == "" {
		return jwtauth.Claims{}, false
	}
	claims, err := jwtauth.Validate(secret, tokenStr)
	if err != nil {
		return jwtauth.Claims{}, false
	}
	return claims, true
}
