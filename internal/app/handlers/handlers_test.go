package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/helpers"
	"github.com/Erlast/short-url.git/internal/app/storages"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := storages.NewMockURLStorage(ctrl)

	tests := []struct {
		name           string
		id             string
		storageResp    string
		storageErr     error
		expectedStatus int
	}{
		{
			name:           "Valid ID",
			id:             "abc123",
			storageResp:    "https://example.com",
			storageErr:     nil,
			expectedStatus: http.StatusTemporaryRedirect,
		},
		{
			name:           "Non-existent ID",
			id:             "notfound",
			storageResp:    "",
			storageErr:     errors.New("not found"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Deleted ID",
			id:             "deleted123",
			storageResp:    "",
			storageErr:     &helpers.ConflictError{},
			expectedStatus: http.StatusGone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store.EXPECT().GetByID(gomock.Any(), tt.id).Return(tt.storageResp, tt.storageErr)

			req, err := http.NewRequest(http.MethodGet, "/"+tt.id, http.NoBody)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			r := chi.NewRouter()

			r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				GetHandler(context.Background(), w, r, store)
			})

			r.ServeHTTP(rr, req)

			resp := rr.Result()
			err = resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.storageResp, rr.Header().Get("Location"))
			}
		})
	}
}

func TestPostHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := storages.NewMockURLStorage(ctrl)

	logger := zap.NewNop().Sugar()

	conf := &config.Cfg{
		FlagBaseURL: "http://localhost:8080",
	}

	originalURL := "http://example.com"
	reqBody := []byte(originalURL)

	store.EXPECT().
		SaveURL(gomock.Any(), "http://example.com").
		Return("newShortURL", nil)

	req, err := http.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := chi.NewRouter()

	r.Post("/shorten", func(w http.ResponseWriter, r *http.Request) {
		PostHandler(context.Background(), w, r, store, conf, logger)
	})

	r.ServeHTTP(rr, req)

	resp := rr.Result()
	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	respBody := rr.Body.String()
	expectedURL := "http://localhost:8080/newShortURL"
	assert.Equal(t, expectedURL, respBody)
}

func TestPostShortenHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := storages.NewMockURLStorage(ctrl)
	conf := &config.Cfg{FlagBaseURL: "http://localhost:8080"}
	logger := zap.NewExample().Sugar()

	tests := []struct {
		name           string
		requestBody    interface{}
		storageResp    string
		storageErr     error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid URL",
			requestBody:    BodyRequested{URL: "https://example.com"},
			storageResp:    "abc123",
			storageErr:     nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"result":"http://localhost:8080/abc123"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatal(err)
			}

			bodyRequested, ok := tt.requestBody.(BodyRequested)
			if !ok {
				t.Fatalf("expected tt.requestBody to be of type BodyRequested, but got %T", tt.requestBody)
			}

			store.EXPECT().SaveURL(gomock.Any(), bodyRequested.URL).Return(tt.storageResp, tt.storageErr)

			rr := httptest.NewRecorder()
			r := chi.NewRouter()

			r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
				PostShortenHandler(context.Background(), w, r, store, conf, logger)
			})

			r.ServeHTTP(rr, req)

			resp := rr.Result()
			err = resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			t.Log(rr.Body.String())
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())
		})
	}
}

func TestBatchShortenHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := storages.NewMockURLStorage(ctrl)
	conf := &config.Cfg{FlagBaseURL: "http://localhost:8080"}
	logger := zap.NewExample().Sugar()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockBehavior   func(store *storages.MockURLStorage, reqBody interface{})
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Empty Request Body",
			requestBody:    nil,
			mockBehavior:   func(store *storages.MockURLStorage, reqBody interface{}) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Empty String!\n",
		},
		{
			name: "Valid Request",
			requestBody: []storages.Incoming{
				{CorrelationID: "3", OriginalURL: "https://example2.com"},
				{CorrelationID: "4", OriginalURL: "https://test.com"},
			},
			mockBehavior: func(store *storages.MockURLStorage, reqBody interface{}) {
				store.EXPECT().LoadURLs(gomock.Any(), reqBody, conf.FlagBaseURL).Return(
					[]storages.Output{
						{CorrelationID: "3", ShortURL: "http://localhost:8080/abc123"},
						{CorrelationID: "4", ShortURL: "http://localhost:8080/def456"},
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: `[
								{"correlation_id":"3","short_url":"http://localhost:8080/abc123"},
								{"correlation_id":"4","short_url":"http://localhost:8080/def456"}
							]`,
		},
		{
			name: "Conflict Error",
			requestBody: []storages.Incoming{
				{CorrelationID: "1", OriginalURL: "https://example.com"},
			},
			mockBehavior: func(store *storages.MockURLStorage, reqBody interface{}) {
				store.EXPECT().LoadURLs(gomock.Any(), reqBody, conf.FlagBaseURL).Return(
					nil,
					&helpers.ConflictError{ShortURL: "someurl", Err: errors.New("someUrl")},
				)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   "",
		},
		{
			name: "Internal Server Error on LoadURLs",
			requestBody: []storages.Incoming{
				{CorrelationID: "1", OriginalURL: "https://example.com"},
			},
			mockBehavior: func(store *storages.MockURLStorage, reqBody interface{}) {
				store.EXPECT().LoadURLs(gomock.Any(), reqBody, conf.FlagBaseURL).Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if tt.requestBody != nil {
				reqBody, _ = json.Marshal(tt.requestBody)
			}

			req, err := http.NewRequest(http.MethodPost, "/api/batch/shorten", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatal(err)
			}

			tt.mockBehavior(store, tt.requestBody)

			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Post("/api/batch/shorten", func(w http.ResponseWriter, r *http.Request) {
				BatchShortenHandler(context.Background(), w, r, store, conf, logger)
			})

			r.ServeHTTP(rr, req)

			resp := rr.Result()
			err = resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedBody != "\n" && tt.expectedBody != "Empty String!\n" && tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			} else {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestGetUserUrls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := storages.NewMockURLStorage(ctrl)
	conf := &config.Cfg{FlagBaseURL: "http://localhost:8080"}
	logger := zap.NewExample().Sugar()

	t.Run("success", func(t *testing.T) {
		expectedResult := []storages.UserURLs{
			{ShortURL: "http://localhost:8080/abc123", OriginalURL: "https://example.com"},
		}

		store.EXPECT().GetUserURLs(gomock.Any(), conf.FlagBaseURL).Return(expectedResult, nil)

		req, err := http.NewRequest(http.MethodGet, "/api/user/urls", http.NoBody)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Get("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(req.Context(), helpers.UserID, 1)
			GetUserUrls(ctx, rr, req, store, conf, logger)
		})

		r.ServeHTTP(rr, req)

		resp := rr.Result()
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var actualResult []storages.UserURLs

		actualResult = append(actualResult, storages.UserURLs{
			ShortURL:    "http://localhost:8080/abc123",
			OriginalURL: "https://example.com",
		})
		err = json.NewDecoder(resp.Body).Decode(&actualResult)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, actualResult)
	})

	t.Run("no content", func(t *testing.T) {
		store.EXPECT().GetUserURLs(gomock.Any(), conf.FlagBaseURL).Return(nil, nil)

		req, err := http.NewRequest(http.MethodGet, "/api/user/urls", http.NoBody)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Get("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), helpers.UserID, 1)
			GetUserUrls(ctx, w, r, store, conf, logger)
		})

		r.ServeHTTP(rr, req)

		resp := rr.Result()
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("storage error", func(t *testing.T) {
		store.EXPECT().GetUserURLs(gomock.Any(), conf.FlagBaseURL).Return(nil, errors.New("storage error"))

		req, err := http.NewRequest(http.MethodGet, "/api/user/urls", http.NoBody)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Get("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(req.Context(), helpers.UserID, 1)
			GetUserUrls(ctx, rr, req, store, conf, logger)
		})

		r.ServeHTTP(rr, req)

		resp := rr.Result()
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestDeleteUserUrls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок для URLStorage
	store := storages.NewMockURLStorage(ctrl)
	logger := zap.NewExample().Sugar()

	t.Run("success", func(t *testing.T) {
		urlsToDelete := []string{"short-url-1", "short-url-2"}
		body, err := json.Marshal(urlsToDelete)
		if err != nil {
			t.Fatal(err)
		}

		store.EXPECT().DeleteUserURLs(gomock.Any(), urlsToDelete, logger).Return(nil)

		req, err := http.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Delete("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), helpers.UserID, 1)
			DeleteUserUrls(ctx, w, r, store, logger)
		})
		r.ServeHTTP(rr, req)

		resp := rr.Result()
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})

	t.Run("empty body", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "/api/user/urls", http.NoBody)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		r := chi.NewRouter()
		r.Delete("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), helpers.UserID, 1)
			DeleteUserUrls(ctx, w, r, store, logger)
		})
		r.ServeHTTP(rr, req)

		resp := rr.Result()
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Empty Body!")
	})

	t.Run("json decode error", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBufferString("invalid-json"))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		DeleteUserUrls(req.Context(), rr, req, store, logger)

		resp := rr.Result()
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("delete error", func(t *testing.T) {
		urlsToDelete := []string{"short-url-1", "short-url-2"}
		body, err := json.Marshal(urlsToDelete)
		if err != nil {
			t.Fatal(err)
		}

		store.EXPECT().DeleteUserURLs(gomock.Any(), urlsToDelete, logger).Return(errors.New("delete failed"))

		req, err := http.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		DeleteUserUrls(req.Context(), rr, req, store, logger)

		resp := rr.Result()
		err = resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
