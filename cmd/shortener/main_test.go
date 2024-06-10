package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/handlers"
	"github.com/Erlast/short-url.git/internal/app/helpers"
	"github.com/Erlast/short-url.git/internal/app/logger"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

func initTestCfg(t *testing.T) (*config.Cfg, storages.URLStorage, *zap.SugaredLogger) {
	t.Helper()

	conf := &config.Cfg{
		FlagRunAddr: ":8080",
		FlagBaseURL: "http://localhost:8080",
	}
	ctx := context.Background()

	newLogger, err := logger.NewLogger("info")

	if err != nil {
		t.Errorf("failed to initialize test cfg (logger): %v", err)
		return nil, nil, nil
	}
	store, err := storages.NewStorage(ctx, conf, newLogger)

	if err != nil {
		t.Errorf("failed to initialize test cfg (storage): %v", err)
		return nil, nil, nil
	}

	return conf, store, newLogger
}

func TestOkPostHandler(t *testing.T) {
	conf, store, newLogger := initTestCfg(t)

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
			ctx := context.Background()
			request := httptest.NewRequest(http.MethodPost, tt.URL, bytes.NewBufferString(tt.body))

			request.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()

			if tt.funcName == "PostHandler" {
				handlers.PostHandler(ctx, w, request, store, conf, newLogger)
			}

			if tt.funcName == "PostShortenHandler" {
				handlers.PostShortenHandler(ctx, w, request, store, conf, newLogger)
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
	conf, store, newLogger := initTestCfg(t)
	ctx := context.Background()

	request := httptest.NewRequest(http.MethodPost, "/", http.NoBody)

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	handlers.PostHandler(ctx, w, request, store, conf, newLogger)

	res := w.Result()

	err := res.Body.Close()

	if err != nil {
		t.Error("Something went wrong")
	}

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestGetHandler(t *testing.T) {
	rndString := helpers.RandomString(7)

	_, store, _ := initTestCfg(t)
	ctx := context.Background()

	err := store.SaveURL(ctx, rndString, "http://somelink.ru")

	if err != nil {
		t.Errorf("unable to save url")
	}

	router := chi.NewRouter()

	handleGet := func(res http.ResponseWriter, req *http.Request) {
		handlers.GetHandler(ctx, res, req, store)
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
	ctx := context.Background()
	rndString := helpers.RandomString(7)

	_, store, _ := initTestCfg(t)

	err := store.SaveURL(ctx, helpers.RandomString(7), "http://somelink.ru")
	if err != nil {
		t.Errorf("unable to save url")
	}

	request := httptest.NewRequest(http.MethodGet, "/"+rndString, http.NoBody)

	request.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	handlers.GetHandler(ctx, w, request, store)

	res := w.Result()

	err = res.Body.Close()

	if err != nil {
		t.Error("Something went wrong")
	}

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestEmptyBodyPostJSONHandler(t *testing.T) {
	conf, store, newLogger := initTestCfg(t)

	ctx := context.Background()

	body := ``

	request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))

	request.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handlers.PostShortenHandler(ctx, w, request, store, conf, newLogger)

	res := w.Result()

	err := res.Body.Close()

	if err != nil {
		t.Error("Something went wrong")
	}

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
