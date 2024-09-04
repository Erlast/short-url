package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Erlast/short-url.git/internal/app/handlers"
	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

func TestNewRouter(t *testing.T) {
	ctx := context.Background()

	conf := &config.Cfg{FlagBaseURL: "http://localhost:8080"}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := storages.NewMockURLStorage(ctrl)

	var logBuf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&logBuf),
		zap.InfoLevel,
	)
	logger := zap.New(core).Sugar()

	r := NewRouter(ctx, store, conf, logger)

	// tests := []struct {
	//	name           string
	//	method         string
	//	id             string
	//	url            string
	//	expectedStatus int
	//	expectedBody   string
	// }{

	//{
	//	name:           "DELETE /api/user/urls",
	//	method:         http.MethodDelete,
	//	url:            "/api/user/urls",
	//	expectedStatus: http.StatusOK,
	//	expectedBody:   "User URLs Deleted", // Настроить в зависимости от вашего обработчика
	// },
	//}

	// for _, tt := range tests {
	t.Run("GET /", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "")

		logs := logBuf.String()
		assert.Contains(t, logs, `status`)
	})

	t.Run("GET /test-id", func(t *testing.T) {
		store.EXPECT().GetByID(gomock.Any(), "test-id").Return("someresp", nil)
		req := httptest.NewRequest(http.MethodGet, "/test-id", http.NoBody)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
		assert.Contains(t, rec.Body.String(), "Temporary Redirect")

		logs := logBuf.String()
		assert.Contains(t, logs, `status`)
	})

	t.Run("POST /", func(t *testing.T) {
		store.EXPECT().
			SaveURL(gomock.Any(), "http://example.com").
			Return("newShortURL", nil)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("http://example.com")))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "newShortURL")

		logs := logBuf.String()
		assert.Contains(t, logs, `status`)
	})

	t.Run("POST /api/shorten", func(t *testing.T) {
		store.EXPECT().SaveURL(gomock.Any(), "https://example.com").Return("abc123", nil)
		reqBody, _ := json.Marshal(handlers.BodyRequested{URL: "https://example.com"})

		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(reqBody))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "abc123")

		logs := logBuf.String()
		assert.Contains(t, logs, `status`)
	})

	t.Run("POST /api/shorten/batch", func(t *testing.T) {
		body := []storages.Incoming{
			{CorrelationID: "3", OriginalURL: "https://example2.com"},
			{CorrelationID: "4", OriginalURL: "https://test.com"},
		}
		reqBody, _ := json.Marshal(body)

		store.EXPECT().LoadURLs(gomock.Any(), body, conf.FlagBaseURL).Return(
			[]storages.Output{
				{CorrelationID: "3", ShortURL: "http://localhost:8080/abc123"},
				{CorrelationID: "4", ShortURL: "http://localhost:8080/def456"},
			}, nil)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBuffer(reqBody))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.JSONEq(t, `[
								{"correlation_id":"3","short_url":"http://localhost:8080/abc123"},
								{"correlation_id":"4","short_url":"http://localhost:8080/def456"}
							]`, rec.Body.String())

		logs := logBuf.String()
		assert.Contains(t, logs, `status`)
	})

	// t.Run("GET /api/user/urls", func(t *testing.T) {
	//	expectedResult := []storages.UserURLs{
	//		{ShortURL: "http://localhost:8080/abc123", OriginalURL: "https://example.com"},
	//	}
	//	store.EXPECT().GetUserURLs(gomock.Any(), conf.FlagBaseURL).Return(expectedResult, nil)
	//	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", http.NoBody)
	//	req.Header.Set("Authorization", "Bearer abc123")
	//	rec := httptest.NewRecorder()
	//	r.ServeHTTP(rec, req)
	//
	//	assert.Equal(t, http.StatusOK, rec.Code)
	//
	//	var actualResult []storages.UserURLs
	//
	//	actualResult = append(actualResult, storages.UserURLs{
	//		ShortURL:    "http://localhost:8080/abc123",
	//		OriginalURL: "https://example.com",
	//	})
	//	err := json.NewDecoder(rec.Body).Decode(&actualResult)
	//	assert.NoError(t, err)
	//	assert.Equal(t, expectedResult, actualResult)
	//
	//	logs := logBuf.String()
	//	assert.Contains(t, logs, `status`)
	// })

	t.Run("DELETE /api/user/urls", func(t *testing.T) {
		urlsToDelete := []string{"short-url-1", "short-url-2"}
		body, err := json.Marshal(urlsToDelete)
		if err != nil {
			t.Fatal(err)
		}

		store.EXPECT().DeleteUserURLs(gomock.Any(), urlsToDelete, logger).Return(nil)
		req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer abc123")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusAccepted, rec.Code)

		logs := logBuf.String()
		assert.Contains(t, logs, `status`)
	})
}
