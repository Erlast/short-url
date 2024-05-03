package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Erlast/short-url.git/internal/app"
	"github.com/Erlast/short-url.git/internal/config"
)

func TestOkPostHandler(t *testing.T) {
	conf := config.ParseFlags()

	app.Init(app.Settings{
		Storage: make(map[string]string),
		Conf:    conf,
	})

	body := "http://somelink.ru"

	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	app.PostHandler(w, request)

	res := w.Result()

	err := res.Body.Close()

	if err != nil {
		t.Error("Something went wrong")
	}

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	resBody, err := io.ReadAll(res.Body)

	require.NoError(t, err)
	assert.NotEmpty(t, string(resBody))
	assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
}

func TestEmptyBodyPostHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", http.NoBody)

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	app.PostHandler(w, request)

	res := w.Result()

	err := res.Body.Close()

	if err != nil {
		t.Error("Something went wrong")
	}

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestGetHandler(t *testing.T) {
	rndString := app.RandomString(7)

	app.Init(app.Settings{
		Storage: map[string]string{rndString: "http://somelink.ru"},
	})

	router := chi.NewRouter()

	router.Get("/{id}", app.GetHandler)

	request := httptest.NewRequest(http.MethodGet, "/"+rndString, http.NoBody)

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, "http://somelink.ru", w.Header().Get("Location"))
}

func TestNotFoundGetHandler(t *testing.T) {
	rndString := app.RandomString(7)

	app.Init(app.Settings{
		Storage: map[string]string{app.RandomString(7): "http://somelink.ru"},
	})

	request := httptest.NewRequest(http.MethodGet, "/"+rndString, http.NoBody)

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	app.GetHandler(w, request)

	res := w.Result()

	err := res.Body.Close()

	if err != nil {
		t.Error("Something went wrong")
	}

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
