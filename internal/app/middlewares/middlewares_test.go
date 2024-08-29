package middlewares

import (
	"bytes"
	"compress/gzip"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/helpers"
	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCheckAuthMiddleware(t *testing.T) {
	logger := zap.NewNop().Sugar()

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
		assert.Equal(t, "Access denied\n", resp.Body.String())
	})

	t.Run("Error getting token from cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "someToken")
		resp := httptest.NewRecorder()

		handler := CheckAuthMiddleware(nextHandler, logger)
		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("Authorization header does not match token cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "someToken")
		req.AddCookie(&http.Cookie{Name: "token", Value: "differentToken"})
		resp := httptest.NewRecorder()

		handler := CheckAuthMiddleware(nextHandler, logger)
		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		assert.Equal(t, "Access denied\n", resp.Body.String())
	})

	t.Run("Authorization header matches token cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.Header.Set("Authorization", "matchingToken")
		req.AddCookie(&http.Cookie{Name: "token", Value: "matchingToken"})
		resp := httptest.NewRecorder()

		handler := CheckAuthMiddleware(nextHandler, logger)
		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "Success", resp.Body.String())
	})
}

func TestGzipMiddleware(t *testing.T) {
	logger := zap.NewNop().Sugar()

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

func TestAuthMiddleware(t *testing.T) {
	logger := zap.NewExample().Sugar()
	cfg := &config.Cfg{}

	t.Run("No token cookie, valid new token", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(helpers.UserID).(string)
			assert.True(t, ok)
			_, err := uuid.Parse(userID)
			assert.NoError(t, err, "userID should be a valid UUID")
			w.WriteHeader(http.StatusOK)
		})

		middleware := AuthMiddleware(handler, logger, cfg)
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		resp := rr.Result()
		err := resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotEmpty(t, rr.Header().Get("Authorization"))
		assert.NotEmpty(t, resp.Cookies())
	})

	t.Run("Invalid token", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := AuthMiddleware(handler, logger, cfg)
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		req.AddCookie(&http.Cookie{Name: "token", Value: "invalid_token"})
		resp := httptest.NewRecorder()

		middleware.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}

func TestWithLogging(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedStatus int
		expectedSize   int
	}{
		{
			name:           "200 OK",
			statusCode:     http.StatusOK,
			responseBody:   "Hello, world!",
			expectedStatus: http.StatusOK,
			expectedSize:   len("Hello, world!"),
		},
		{
			name:           "404 Not Found",
			statusCode:     http.StatusNotFound,
			responseBody:   "Page not found",
			expectedStatus: http.StatusNotFound,
			expectedSize:   len("Page not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf bytes.Buffer

			// Настраиваем zap для записи логов в буфер.
			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
				zapcore.AddSync(&logBuf),
				zap.InfoLevel,
			)
			logger := zap.New(core).Sugar()

			handler := WithLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}), logger)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			// Проверяем статус и размер ответа.
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.responseBody, rec.Body.String())
		})
	}
}
