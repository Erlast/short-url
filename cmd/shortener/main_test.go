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

func initTestCfg() (*config.Cfg, *storages.Storage) {
	conf := &config.Cfg{
		FlagRunAddr: ":8080",
		FlagBaseURL: "http://localhost:8080",
	}

	store := storages.NewStorage()

	return conf, store
}

func TestOkPostHandler(t *testing.T) {
	conf, store := initTestCfg()

	tests := []struct {
		name        string
		body        string
		URL         string
		contentType string
		funcName    string
	}{
		{
			name:        "ok post",
			body:        "http://somelink.ru",
			URL:         "/",
			contentType: "text/plain",
			funcName:    "PostHandler",
		},
		{
			name:        "ok post /api/shorten",
			body:        `{"url": "http://somelink.ru"}`,
			URL:         "/api/shorten",
			contentType: "application/json",
			funcName:    "PostShortenHandler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.URL, bytes.NewBufferString(tt.body))

			request.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()

			if tt.funcName == "PostHandler" {
				handlers.PostHandler(w, request, store, conf)
			}

			if tt.funcName == "PostShortenHandler" {
				handlers.PostShortenHandler(w, request, store, conf)
			}

			res := w.Result()

			err := res.Body.Close()

			if err != nil {
				t.Error("Something went wrong")
			}

			assert.Equal(t, http.StatusCreated, res.StatusCode)
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.NotEmpty(t, string(resBody))
			assert.Equal(t, tt.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestEmptyBodyPostHandler(t *testing.T) {
	conf, store := initTestCfg()

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
func TestEmptyBodyPostJSONHandler(t *testing.T) {
	conf, store := initTestCfg()

	body := ``

	request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))

	request.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handlers.PostShortenHandler(w, request, store, conf)

	res := w.Result()

	err := res.Body.Close()

	if err != nil {
		t.Error("Something went wrong")
	}

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
