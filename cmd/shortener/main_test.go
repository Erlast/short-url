package main

import (
	"bytes"
	"io"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOkPostHandler(t *testing.T) {

	body := "http://somelink.ru"

	storage = make(map[string]string)

	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(body)))

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	postHandler(w, request)

	res := w.Result()

	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	resBody, err := io.ReadAll(res.Body)

	require.NoError(t, err)
	assert.NotEmpty(t, string(resBody))
	assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))

}

func TestEmptyBodyPostHandler(t *testing.T) {

	request := httptest.NewRequest(http.MethodPost, "/", nil)

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	postHandler(w, request)

	res := w.Result()

	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

}

func TestGetHandler(t *testing.T) {
	storage = make(map[string]string)

	rndString := randomString(7)

	storage[rndString] = "http://somelink.ru"

	request := httptest.NewRequest(http.MethodGet, "/"+rndString, nil)

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	getHandler(w, request)

	res := w.Result()

	defer res.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.Equal(t, "http://somelink.ru", res.Header.Get("Location"))
}

func TestNotFoundGetHandler(t *testing.T) {
	storage = make(map[string]string)

	rndString := randomString(7)

	storage[randomString(7)] = "http://somelink.ru"

	request := httptest.NewRequest(http.MethodGet, "/"+rndString, nil)

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	getHandler(w, request)

	res := w.Result()

	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
