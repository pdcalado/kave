package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testAddress = "http://localhost:8000"
)

func waitUntilHealthy(t *testing.T) {
	for i := 0; i < 20; i++ {
		res, err := http.Get(testAddress + defaultHealthPath)
		if err != nil || res.StatusCode/100 != 2 {
			time.Sleep(time.Millisecond * 50)
			continue
		}
		return
	}

	t.Fatalf("failed to start server")
}

func TestMain(t *testing.T) {
	// start the server
	configFile, err := os.CreateTemp(os.TempDir(), "config")
	assert.NoError(t, err)

	_, err = io.WriteString(configFile, `
address = "localhost:8000"
redis_address = "localhost:6379"
	`)
	assert.NoError(t, err)

	go run(configFile.Name())

	waitUntilHealthy(t)

	testKey := "foo"

	// set a key
	res, err := http.Post(
		testAddress+defaultRouterBasePath+"/"+testKey,
		"application/json",
		bytes.NewBufferString(`{}`),
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	// get the key
	res, err = http.Get(testAddress + defaultRouterBasePath + "/" + testKey)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	buf, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, `{}`, string(buf))

	// get a non existing key
	res, err = http.Get(testAddress + defaultRouterBasePath + "/notfound")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
