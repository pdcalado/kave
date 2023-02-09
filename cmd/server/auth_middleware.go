package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type AuthMiddleware struct {
	parseToken      parseTokenFunc
	permissionToCtx func(ctx context.Context, permissions []string) context.Context
}

// parseTokenFunc parses the Authorization token and returns a list of permissions.
type parseTokenFunc func(string) ([]string, error)

func NewAuthMiddleware(
	parseToken parseTokenFunc,
	permissionToCtx func(ctx context.Context, permissions []string) context.Context,
) AuthMiddleware {
	return AuthMiddleware{
		parseToken:      parseToken,
		permissionToCtx: permissionToCtx,
	}
}

type claimsWithPermissions struct {
	jwt.RegisteredClaims
	Permissions []string `json:"permissions"`
}

func (m AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := extractTokenFromHeaders(r)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		permissions, err := m.parseToken(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		newContext := m.permissionToCtx(r.Context(), permissions)

		next.ServeHTTP(w, r.WithContext(newContext))
	})
}

// extractTokenFromHeaders looks for authorization header and gets token
func extractTokenFromHeaders(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", false
	}

	parts := strings.Split(authHeader, " ")
	if !strings.EqualFold(parts[0], "bearer") {
		return "", false
	}

	return parts[1], true
}
