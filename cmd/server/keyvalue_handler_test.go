package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	mocks "github.com/pdcalado/kave/cmd/server/mocks"
)

type mockResponseWriter struct {
	t            *testing.T
	expectedCode int
	expectedBody []byte
	expectWrite  bool
}

func (m *mockResponseWriter) WriteHeader(code int) {
	assert.Equal(m.t, m.expectedCode, code)
}

func (m *mockResponseWriter) Write(body []byte) (int, error) {
	assert.True(m.t, m.expectWrite)

	assert.Equal(m.t, m.expectedBody, body)
	return len(body), nil
}

func (m *mockResponseWriter) Header() http.Header {
	return http.Header{}
}

func TestKeyValueHandler(t *testing.T) {
	// get a key that exists
	{
		ctx := context.Background()

		testKey := "foo"

		kv := mocks.NewMockKeyValue(gomock.NewController(t))

		handler := NewKeyValueHandler(kv, "prefix:", func(ctx context.Context) string {
			return "foo"
		})

		kv.EXPECT().Get(gomock.Any(), "prefix:"+testKey).Return("value", nil)

		request, _ := http.NewRequestWithContext(context.WithValue(ctx, redisKey{}, testKey), http.MethodGet, "http://localhost:8080", nil)

		handler.Get(
			&mockResponseWriter{
				t:            t,
				expectedCode: http.StatusOK,
				expectedBody: []byte("value"),
				expectWrite:  true,
			},
			request,
		)
	}

	// get a key that does not exist
	{
		ctx := context.Background()

		testKey := "foo"

		kv := mocks.NewMockKeyValue(gomock.NewController(t))

		handler := NewKeyValueHandler(kv, "prefix:", func(ctx context.Context) string {
			return "foo"
		})

		kv.EXPECT().Get(gomock.Any(), "prefix:"+testKey).Return("", ErrorKeyNotFound{})

		request, _ := http.NewRequestWithContext(context.WithValue(ctx, redisKey{}, testKey), http.MethodGet, "http://localhost:8080", nil)

		handler.Get(
			&mockResponseWriter{
				t:            t,
				expectedCode: http.StatusNotFound,
				expectWrite:  false,
			},
			request,
		)
	}

	// get a key and fail to obtain it
	{
		ctx := context.Background()

		testKey := "foo"

		kv := mocks.NewMockKeyValue(gomock.NewController(t))

		handler := NewKeyValueHandler(kv, "prefix:", func(ctx context.Context) string {
			return "foo"
		})

		kv.EXPECT().Get(gomock.Any(), "prefix:"+testKey).Return("", fmt.Errorf("something went wrong"))

		request, _ := http.NewRequestWithContext(context.WithValue(ctx, redisKey{}, testKey), http.MethodGet, "http://localhost:8080", nil)

		handler.Get(
			&mockResponseWriter{
				t:            t,
				expectedCode: http.StatusInternalServerError,
				expectWrite:  false,
			},
			request,
		)
	}
}
