package main

import (
	"bytes"
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

type failingReader struct{}

func (f *failingReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("failed to write")
}

func TestKeyValueHandlerGet(t *testing.T) {
	// get a key that exists
	{
		ctx := context.Background()

		testKey := "foo"

		kv := mocks.NewMockKeyValue(gomock.NewController(t))

		handler := NewKeyValueHandler(kv, "prefix:", func(ctx context.Context) string {
			return testKey
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
			return testKey
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
			return testKey
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

func TestKeyValueHandlerSet(t *testing.T) {
	// set a key successfully exists
	{
		ctx := context.Background()

		testKey := "foo"

		kv := mocks.NewMockKeyValue(gomock.NewController(t))

		handler := NewKeyValueHandler(kv, "prefix:", func(ctx context.Context) string {
			return testKey
		})

		kv.EXPECT().Set(gomock.Any(), "prefix:"+testKey, []byte("value")).Return(nil)

		request, _ := http.NewRequestWithContext(context.WithValue(ctx, redisKey{}, testKey), http.MethodPost, "http://localhost:8080", bytes.NewBuffer([]byte("value")))

		handler.Set(
			&mockResponseWriter{
				t:            t,
				expectedCode: http.StatusCreated,
			},
			request,
		)
	}

	// set a key with empty body
	{
		ctx := context.Background()

		testKey := "foo"

		kv := mocks.NewMockKeyValue(gomock.NewController(t))

		handler := NewKeyValueHandler(kv, "prefix:", func(ctx context.Context) string {
			return testKey
		})

		request, _ := http.NewRequestWithContext(context.WithValue(ctx, redisKey{}, testKey), http.MethodPost, "http://localhost:8080", &failingReader{})

		handler.Set(
			&mockResponseWriter{
				t:            t,
				expectedCode: http.StatusInternalServerError,
				expectWrite:  false,
			},
			request,
		)
	}

	// set a key and fail on backend client
	{
		ctx := context.Background()

		testKey := "foo"

		kv := mocks.NewMockKeyValue(gomock.NewController(t))

		handler := NewKeyValueHandler(kv, "prefix:", func(ctx context.Context) string {
			return testKey
		})

		kv.EXPECT().Set(gomock.Any(), "prefix:"+testKey, []byte("value")).Return(fmt.Errorf("something went wrong"))

		request, _ := http.NewRequestWithContext(context.WithValue(ctx, redisKey{}, testKey), http.MethodGet, "http://localhost:8080", bytes.NewBuffer([]byte("value")))

		handler.Set(
			&mockResponseWriter{
				t:            t,
				expectedCode: http.StatusInternalServerError,
				expectWrite:  false,
			},
			request,
		)
	}
}
