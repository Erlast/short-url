package middlewares

import (
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const emptyStatus = 0

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write переопределяем метод Write запроса.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return 0, errors.New("error writing response")
	}
	r.responseData.size += size
	if r.responseData.status == emptyStatus {
		r.responseData.status = http.StatusOK
	}
	return size, nil
}

// WriteHeader переопределяем метод WriteHeader запроса.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// Header переопредеялем метод Header запроса.
func (r *loggingResponseWriter) Header() http.Header {
	return r.ResponseWriter.Header()
}

// WithLogging функция логгирования http запросов.
func WithLogging(h http.Handler, logger *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	})
}
