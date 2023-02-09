package main

import (
	"context"
	"io"
	"log"
	"net/http"
)

type KeyValue interface {
	Get(context.Context, string) (string, error)
	Set(context.Context, string, []byte) error
}

// create a test function for this struct
type KeyValueHandler struct {
	client         KeyValue
	prefix         string
	keyFromContext func(context.Context) string
}

type ErrorKeyNotFound struct{}

func (e ErrorKeyNotFound) Error() string {
	return "key not found"
}

func (e ErrorKeyNotFound) Is(target error) bool {
	_, ok := target.(ErrorKeyNotFound)
	return ok
}

func NewKeyValueHandler(
	client KeyValue,
	prefix string,
	keyFromContext func(context.Context) string,
) *KeyValueHandler {
	return &KeyValueHandler{
		client:         client,
		prefix:         prefix,
		keyFromContext: keyFromContext,
	}
}

func (kv *KeyValueHandler) formatKey(key string) string {
	return kv.prefix + key
}

func (kv *KeyValueHandler) Get(w http.ResponseWriter, r *http.Request) {
	key := kv.formatKey(kv.keyFromContext(r.Context()))

	value, err := kv.client.Get(r.Context(), key)
	if (ErrorKeyNotFound{}).Is(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error getting key %s: %v\n", key, err)
		return
	}

	_, err = w.Write([]byte(value))
	if err != nil {
		log.Printf("error writing response: %v\n", err)
		return
	}
}

func (kv *KeyValueHandler) Set(w http.ResponseWriter, r *http.Request) {
	key := kv.formatKey(kv.keyFromContext(r.Context()))

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error reading body: %v", err)
		return
	}

	err = kv.client.Set(r.Context(), key, body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error getting key %s: %v", key, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
