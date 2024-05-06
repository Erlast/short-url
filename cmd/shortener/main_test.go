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

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/handlers"
	"github.com/Erlast/short-url.git/internal/app/helpers"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

func TestOkPostHandler(t *testing.T) {
	conf := config.ParseFlags()

	store := storages.NewStorage()

	body := "http://somelink.ru"

	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	handlers.PostHandler(w, request, store, conf)

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
	conf := config.ParseFlags()

	store := storages.NewStorage()

	request := httptest.NewRequest(http.MethodPost, "/", http.NoBody)

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	handlers.PostHandler(w, request, store, conf)

	res := w.Result()

	err := res.Body.Close()

	if err != nil {
		t.Error("Something went wrong")
	}

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestGetHandler(t *testing.T) {
	rndString := helpers.RandomString(7)

	store := storages.NewStorage()

	store.SaveURL(rndString, "http://somelink.ru")

	router := chi.NewRouter()

	handleGet := func(res http.ResponseWriter, req *http.Request) {
		handlers.GetHandler(res, req, store)
	}

	router.Get("/{id}", handleGet)

	request := httptest.NewRequest(http.MethodGet, "/"+rndString, http.NoBody)

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	assert.Equal(t, "http://somelink.ru", w.Header().Get("Location"))
}

func TestNotFoundGetHandler(t *testing.T) {
	rndString := helpers.RandomString(7)

	store := storages.NewStorage()

	store.SaveURL(helpers.RandomString(7), "http://somelink.ru")

	request := httptest.NewRequest(http.MethodGet, "/"+rndString, http.NoBody)

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	handlers.GetHandler(w, request, store)

	res := w.Result()

	err := res.Body.Close()

	if err != nil {
		t.Error("Something went wrong")
	}

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
