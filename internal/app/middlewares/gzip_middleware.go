package middlewares

import (
	"net/http"
	"strings"

	"github.com/Erlast/short-url.git/internal/app/logger"
	"github.com/Erlast/short-url.git/internal/app/zip"
)

func GzipMiddleware(h http.Handler) http.Handler {
	zipFn := func(resp http.ResponseWriter, req *http.Request) {
		ow := resp

		contentType := req.Header.Get("Content-Type")
		supportsContentTypeText := strings.Contains(contentType, "text/plain")
		supportsContentTypeJSON := strings.Contains(contentType, "application/json")

		if !supportsContentTypeText && !supportsContentTypeJSON {
			h.ServeHTTP(resp, req)
			return
		}

		acceptEncoding := req.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := zip.NewCompressWriter(resp)
			ow = cw
			err := cw.Close()
			if err != nil {
				logger.Log.Errorln(err)
			}
		}

		contentEncoding := req.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := zip.NewCompressReader(req.Body)
			if err != nil {
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Body = cr
			defer func(cr *zip.CompressReader) {
				err := cr.Close()
				if err != nil {
					logger.Log.Errorln(err)
				}
			}(cr)
		}

		h.ServeHTTP(ow, req)
	}

	return http.HandlerFunc(zipFn)
}
