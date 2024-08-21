package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCheckAuthMiddleware(t *testing.T) {
	logger := zap.NewNop().Sugar() // Используем заглушку логгера для тестов

	// Следующий обработчик, который будет вызван, если CheckAuthMiddleware пройдёт
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Success"))
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("No Authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		resp := httptest.NewRecorder()

		handler := CheckAuthMiddleware(nextHandler, logger)
		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Equal(t, "Access denied\n", resp.Body.String()) // assuming accessDeniedErr = "access denied"
	})

	t.Run("Error getting token from cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "someToken") // Устанавливаем заголовок Authorization
		resp := httptest.NewRecorder()

		handler := CheckAuthMiddleware(nextHandler, logger)
		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("Authorization header does not match token cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "someToken")                        // Устанавливаем заголовок Authorization
		req.AddCookie(&http.Cookie{Name: "token", Value: "differentToken"}) // Устанавливаем несовпадающий токен
		resp := httptest.NewRecorder()

		handler := CheckAuthMiddleware(nextHandler, logger)
		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Equal(t, "Access denied\n", resp.Body.String()) // assuming accessDeniedErr = "access denied"
	})

	t.Run("Authorization header matches token cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "matchingToken")                   // Устанавливаем заголовок Authorization
		req.AddCookie(&http.Cookie{Name: "token", Value: "matchingToken"}) // Устанавливаем совпадающий токен
		resp := httptest.NewRecorder()

		handler := CheckAuthMiddleware(nextHandler, logger)
		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "Success", resp.Body.String())
	})
}

func TestGzipMiddleware(t *testing.T) {
	logger := zap.NewNop().Sugar() // Используем заглушку логгера для тестов

	// Обработчик, который будет вызван, если GzipMiddleware пропустит запрос
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"message": "Hello, World!"}`))
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Gzip compressed request body", func(t *testing.T) {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, err := gz.Write([]byte("test body"))
		assert.NoError(t, err)
		err = gz.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/", &buf)
		req.Header.Set("Content-Encoding", "gzip")
		resp := httptest.NewRecorder()

		handler := GzipMiddleware(nextHandler, logger)
		handler.ServeHTTP(resp, req)

		body, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, "test body", string(body))
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("Gzip compression for response", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		handler := GzipMiddleware(nextHandler, logger)
		handler.ServeHTTP(resp, req)

		assert.Equal(t, "gzip", resp.Header().Get("Content-Encoding"))

		gr, err := gzip.NewReader(resp.Body)
		assert.NoError(t, err)

		err = gr.Close()
		if err != nil {
			t.Fatal(err)
		}

		body, err := io.ReadAll(gr)
		assert.NoError(t, err)
		assert.Equal(t, `{"message": "Hello, World!"}`, string(body))
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("Request without gzip support", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		resp := httptest.NewRecorder()

		handler := GzipMiddleware(nextHandler, logger)
		handler.ServeHTTP(resp, req)

		assert.NotContains(t, resp.Header().Get("Content-Encoding"), "gzip")
		assert.Equal(t, `{"message": "Hello, World!"}`, resp.Body.String())
		assert.Equal(t, http.StatusOK, resp.Code)
	})
}
