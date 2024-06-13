package gzip

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
)

type CompressWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
	}
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	bytesRead, err := c.zw.Write(p)
	if err != nil {
		return 0, fmt.Errorf("an error occured writing to compressor: %w", err)
	}
	return bytesRead, nil
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < http.StatusMultipleChoices {
		c.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	err := c.zw.Close()
	if err != nil {
		return fmt.Errorf("an error occured closing compressor: %w", err)
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
		return nil, fmt.Errorf("cannot create newCompressReade: %w", err)
	}

	return &CompressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c CompressReader) Read(p []byte) (n int, err error) {
	bytesRead, err := c.zr.Read(p)
	if err != nil {
		return 0, fmt.Errorf("cannot read from compressReader: %w", err)
	}
	if err := c.zr.Close(); err != nil {
		return 0, fmt.Errorf("cannot close compressReader: %w", err)
	}
	return bytesRead, nil
}

func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("cannot close compressReader: %w", err)
	}
	return nil
}
