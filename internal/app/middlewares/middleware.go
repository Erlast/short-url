package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

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

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogging(h http.Handler, logger *zap.Logger) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
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

		logger.Info("got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.String("status", strconv.Itoa(responseData.status)),
			zap.String("duration", strconv.FormatInt(int64(duration), 10)),
			zap.String("size", strconv.Itoa(responseData.size)),
		)
	}
	return http.HandlerFunc(logFn)
}
