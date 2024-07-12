package routes

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MihailSergeenkov/gophermart/internal/app/common"
	"go.uber.org/zap"
)

const maxStatusCode = 300

type compressWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
	}
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p) //nolint:wrapcheck // Нужно обернуть, но возврат должен остаться оригинальным
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < maxStatusCode {
		c.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close() //nolint:wrapcheck // Нужно обернуть, но возврат должен остаться оригинальным
}

type compressReader struct {
	r      io.ReadCloser
	zr     *gzip.Reader
	logger *zap.Logger
}

func newCompressReader(r io.ReadCloser, l *zap.Logger) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to init gzip reader: %w", err)
	}

	return &compressReader{
		r:      r,
		zr:     zr,
		logger: l,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p) //nolint:wrapcheck // Нужно обернуть, но возврат должен остаться оригинальным
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		c.logger.Error("failed to close base reader", zap.Error(err))
	}
	return c.zr.Close() //nolint:wrapcheck // Нужно обернуть, но возврат должен остаться оригинальным
}

func gzipMiddleware(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				cw := newCompressWriter(w)
				ow = cw
				defer func() {
					err := cw.Close()

					if err != nil {
						l.Error("failed to close compress writer", zap.Error(err))
					}
				}()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				contentType := r.Header.Get(common.ContentTypeHeader)
				if !(strings.Contains(contentType, common.JSONContentType) || strings.Contains(contentType, "text/html")) {
					l.Warn("content encoding for bad content type", zap.String("content_type", contentType))
				}

				cr, err := newCompressReader(r.Body, l)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					l.Error("failed to create compress reader", zap.Error(err))
					return
				}

				r.Body = cr
				defer func() {
					err := cr.Close()

					if err != nil {
						l.Error("failed to close compress reader", zap.Error(err))
					}
				}()
			}

			next.ServeHTTP(ow, r)
		})
	}
}
