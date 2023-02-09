package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"golang.org/x/exp/slices"
)

type PermissionMiddleware struct {
	keyPrefix          string
	matcher            *memoizedMatcher
	keyFromCtx         func(context.Context) string
	permissionsFromCtx func(context.Context) []string
}

func NewPermissionMiddleware(
	keyPrefix string,
	keyFromCtx func(context.Context) string,
	permissionsFromCtx func(context.Context) []string,
) PermissionMiddleware {
	return PermissionMiddleware{
		keyPrefix: keyPrefix,
		matcher: &memoizedMatcher{
			inner: make(map[string]*regexp.Regexp),
		},
		keyFromCtx:         keyFromCtx,
		permissionsFromCtx: permissionsFromCtx,
	}
}

type memoizedMatcher struct {
	inner map[string]*regexp.Regexp
}

func (m *memoizedMatcher) MatchString(pattern string, s string) bool {
	r, exists := m.inner[pattern]
	if !exists {
		var err error
		r, err = regexp.Compile(pattern)
		if err != nil {
			log.Printf("failed to compile pattern '%s': %s", pattern, err)
			return false
		}
	}
	return r.MatchString(s)
}

func (p PermissionMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get key from context
		key := p.keyFromCtx(r.Context())
		// get permissions from context
		permissions := p.permissionsFromCtx(r.Context())

		var required string

		switch r.Method {
		case http.MethodGet:
			required = fmt.Sprintf("read:%s%s", p.keyPrefix, key)
		case http.MethodPost:
			required = fmt.Sprintf("write:%s%s", p.keyPrefix, key)
		default:
			w.WriteHeader(http.StatusForbidden)
			return
		}

		index := slices.IndexFunc(permissions, func(pattern string) bool {
			return p.matcher.MatchString(pattern, required)
		})

		if index == -1 {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
