package gzip

import (
	"compress/gzip"
	"fmt"
	logger "go-yandex-metrics/cmd/server/middleware"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	bytesRead, err := c.zw.Write(p)
	if err != nil {
		return 0, fmt.Errorf("an error occured writing to compressor: %w", err)
	}
	return bytesRead, nil
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < http.StatusMultipleChoices {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	err := c.zw.Close()
	if err != nil {
		return fmt.Errorf("an error occured closing compressor: %w", err)
	}
	return nil
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("cannot create newCompressReade: %w", err)
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	bytesRead, err := c.zr.Read(p)
	if err != nil {
		return 0, fmt.Errorf("cannot read from compressReader: %w", err)
	}
	if err := c.zr.Close(); err != nil {
		return 0, fmt.Errorf("cannot close compressReader: %w", err)
	}
	return bytesRead, nil
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("cannot close compressReader: %w", err)
	}
	return nil
}

func GzipMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			lg := logger.InitLogger()

			ow := w
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				cw := newCompressWriter(w)
				ow = cw
				defer func() {
					if err := cw.Close(); err != nil {
						lg.Info(fmt.Sprintf("failed to close newCompressWriter: %v", err))
						return
					}
				}()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					lg.Info(fmt.Sprintf("failed to close newCompressWriter: %v", err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer func() {
					if err := cr.Close(); err != nil {
						lg.Info(fmt.Sprintf("failed to close newCompressReader: %v", err))
						return
					}
				}()
			}

			next.ServeHTTP(ow, r)
		}
		return http.HandlerFunc(fn)
	}
}
