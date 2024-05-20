package zip

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
)

const acceptedStatusCode = 300

type CompressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	size, err := c.zw.Write(p)
	if err != nil {
		return 0, errors.New("error writing response")
	}
	return size, nil
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < acceptedStatusCode {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	err := c.zw.Close()
	if err != nil {
		return errors.New("error close gzip")
	}
	return nil
}

type CompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, errors.New("error reading gzip")
	}

	return &CompressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c CompressReader) Read(p []byte) (n int, err error) {
	size, err := c.zr.Read(p)
	if err != nil {
		return 0, errors.New("error reading compress")
	}
	return size, nil
}

func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return errors.New("error closing reader")
	}

	if err := c.zr.Close(); err != nil {
		return errors.New("error closing compress reader")
	}
	return nil
}
