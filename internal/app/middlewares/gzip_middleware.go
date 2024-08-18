package middlewares

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// GzipResponseWriter структура ответа при сжатии данных
type GzipResponseWriter struct {
	Writer io.Writer
	http.ResponseWriter
}

// Write переопределяеми метод Write пакета gzip
func (w *GzipResponseWriter) Write(b []byte) (int, error) {
	size, err := w.Writer.Write(b)
	if err != nil {
		return 0, errors.New("error writing to gzip writer")
	}
	return size, nil
}

// GzipMiddleware функция сжатия данных запроса.
func GzipMiddleware(h http.Handler, logger *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		contentEncoding := req.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			gReader, err := gzip.NewReader(req.Body)
			if err != nil {
				logger.Errorln(err)
				http.Error(resp, "Invalid body", http.StatusInternalServerError)
				return
			}

			defer func(gReader *gzip.Reader) {
				err := gReader.Close()
				if err != nil {
					logger.Errorln(err)
				}
			}(gReader)

			req.Body = gReader
			h.ServeHTTP(resp, req)
			return
		}

		contentType := req.Header.Get("Content-Type")
		supportsContentTypeText := strings.Contains(contentType, "text/plain")
		supportsContentTypeJSON := strings.Contains(contentType, "application/json")

		isCompressed := supportsContentTypeText || supportsContentTypeJSON

		acceptEncoding := req.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip && isCompressed {
			gWriter := gzip.NewWriter(resp)

			defer func(gWriter *gzip.Writer) {
				err := gWriter.Close()
				if err != nil {
					logger.Errorln(err)
				}
			}(gWriter)

			resp.Header().Set("Content-Encoding", "gzip")

			gzipResponseWriter := &GzipResponseWriter{Writer: gWriter, ResponseWriter: resp}
			h.ServeHTTP(gzipResponseWriter, req)
			return
		}
		h.ServeHTTP(resp, req)
	})
}
