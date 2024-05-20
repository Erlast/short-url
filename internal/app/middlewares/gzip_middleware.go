package middlewares

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Erlast/short-url.git/internal/app/logger"
)

type GzipResponseWriter struct {
	Writer io.Writer
	http.ResponseWriter
}

func (w *GzipResponseWriter) Write(b []byte) (int, error) {
	write, err := w.Writer.Write(b)
	if err != nil {
		return 0, errors.New("error writing to gzip writer")
	}
	return write, nil
}

func GzipMiddleware(h http.Handler) http.Handler {
	zipFn := func(resp http.ResponseWriter, req *http.Request) {
		contentType := req.Header.Get("Content-Type")
		supportsContentTypeText := strings.Contains(contentType, "text/plain")
		supportsContentTypeJSON := strings.Contains(contentType, "application/json")

		if !supportsContentTypeText && !supportsContentTypeJSON {
			h.ServeHTTP(resp, req)
			return
		}

		contentEncoding := req.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			gReader, err := gzip.NewReader(req.Body)
			if err != nil {
				http.Error(resp, "Invalid body", http.StatusInternalServerError)
				return
			}

			defer func(gReader *gzip.Reader) {
				err := gReader.Close()
				if err != nil {
					logger.Log.Errorln(err)
				}
			}(gReader)

			req.Body = gReader
			h.ServeHTTP(resp, req)
			return
		}

		acceptEncoding := req.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			gWriter := gzip.NewWriter(resp)

			defer func(gWriter *gzip.Writer) {
				err := gWriter.Close()
				if err != nil {
					logger.Log.Errorln(err)
				}
			}(gWriter)

			resp.Header().Set("Content-Encoding", "gzip")

			gzipResponseWriter := &GzipResponseWriter{Writer: gWriter, ResponseWriter: resp}
			h.ServeHTTP(gzipResponseWriter, req)
			return
		}
	}

	return http.HandlerFunc(zipFn)
}
