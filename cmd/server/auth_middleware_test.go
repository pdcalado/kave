package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	// test a request without token
	{
		ctx := context.Background()

		token := "token"
		permissions := []string{"read:nothing"}

		parser := func(tkn string) ([]string, error) {
			assert.Equal(t, token, tkn)
			return permissions, nil
		}

		am := NewAuthMiddleware(
			parser,
			func(ctx context.Context, perms []string) context.Context {
				assert.Equal(t, permissions, perms)
				return ctx
			},
		)

		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080", nil)

		am.Handler(nil).ServeHTTP(
			&mockResponseWriter{
				t:            t,
				expectedCode: http.StatusUnauthorized,
			},
			request,
		)
	}

	// test a request without invalid token
	{
		ctx := context.Background()

		token := "token"
		permissions := []string{"read:nothing"}

		parser := func(tkn string) ([]string, error) {
			assert.Equal(t, token, tkn)
			return nil, fmt.Errorf("failed to parse token")
		}

		am := NewAuthMiddleware(
			parser,
			func(ctx context.Context, perms []string) context.Context {
				assert.Equal(t, permissions, perms)
				return ctx
			},
		)

		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080", nil)
		request.Header.Set("Authorization", "Bearer "+token)

		am.Handler(nil).ServeHTTP(
			&mockResponseWriter{
				t:            t,
				expectedCode: http.StatusUnauthorized,
			},
			request,
		)
	}

	// test a request without invalid token
	{
		ctx := context.Background()

		token := "token"
		permissions := []string{"read:nothing"}

		parser := func(tkn string) ([]string, error) {
			assert.Equal(t, token, tkn)
			return permissions, nil
		}

		type foo struct{}

		am := NewAuthMiddleware(
			parser,
			func(ctx context.Context, perms []string) context.Context {
				assert.Equal(t, permissions, perms)
				return context.WithValue(ctx, foo{}, "bar")
			},
		)

		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080", nil)
		request.Header.Set("Authorization", "Bearer "+token)

		next := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "bar", r.Context().Value(foo{}))
		}

		am.Handler(http.HandlerFunc(next)).ServeHTTP(
			nil,
			request,
		)
	}
}
