package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionMiddleware(t *testing.T) {
	// test a forbidden request
	{
		ctx := context.Background()

		pm := NewPermissionMiddleware(
			"prefix:",
			func(ctx context.Context) string {
				return "foo"
			},
			func(ctx context.Context) []string {
				return []string{"read:nothing"}
			},
		)

		handler := pm.Handler(nil)

		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080", nil)

		handler.ServeHTTP(
			&mockResponseWriter{
				t:            t,
				expectedCode: http.StatusForbidden,
			},
			request,
		)
	}

	// test a request with unhandled method
	{
		ctx := context.Background()

		pm := NewPermissionMiddleware(
			"prefix:",
			func(ctx context.Context) string {
				return "foo"
			},
			func(ctx context.Context) []string {
				return []string{"read:nothing"}
			},
		)

		handler := pm.Handler(nil)

		request, _ := http.NewRequestWithContext(ctx, http.MethodDelete, "http://localhost:8080", nil)

		handler.ServeHTTP(
			&mockResponseWriter{
				t:            t,
				expectedCode: http.StatusForbidden,
			},
			request,
		)
	}

	// test a request with allowed permissions
	{
		ctx := context.Background()

		pm := NewPermissionMiddleware(
			"prefix:",
			func(ctx context.Context) string {
				return "foo"
			},
			func(ctx context.Context) []string {
				return []string{"read:*"}
			},
		)

		wasCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wasCalled = true
		})

		handler := pm.Handler(next)

		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080", nil)

		handler.ServeHTTP(
			&mockResponseWriter{
				t:            t,
				expectedCode: http.StatusForbidden,
			},
			request,
		)

		assert.True(t, wasCalled)
	}
}
