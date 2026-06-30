package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"slices"
	"strings"

	models "github.com/scouser-122/go-metrics/internal/model"
)

type gzipWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

var compressibleTypes = []string{
	"text/html",
	"application/json",
}

func newGzipWriter() *gzipWriter {
	return &gzipWriter{
		w:  nil,
		zw: gzip.NewWriter(io.Discard),
	}
}

var gzipWriterPool = models.NewPool(newGzipWriter)

func (c *gzipWriter) Header() http.Header {
	return c.w.Header()
}

func (c *gzipWriter) Write(p []byte) (int, error) {
	if c.shouldCompress() {
		c.zw.Reset(c.w)
		return c.zw.Write(p)
	}
	return c.w.Write(p)
}

func (c *gzipWriter) WriteHeader(statusCode int) {
	if c.shouldCompress() && statusCode < 300 {
		c.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *gzipWriter) shouldCompress() bool {
	contentType := c.w.Header().Get("Content-Type")
	for _, ct := range compressibleTypes {
		if strings.HasPrefix(contentType, ct) {
			return true
		}
	}
	return false
}

func (c *gzipWriter) Close() error {
	if c.zw != nil {
		err := c.zw.Close()
		return err
	}
	return nil
}

func (c *gzipWriter) Reset() {
	c.Close()
	c.w = nil
}

type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newGzipReader() *gzipReader {
	return &gzipReader{
		r:  nil,
		zr: new(gzip.Reader),
	}
}

var gzipReaderPool = models.NewPool(newGzipReader)

func (c *gzipReader) SetUp(r io.ReadCloser) error {
	err := c.zr.Reset(r)
	if err != nil {
		return err
	}
	c.r = r
	return nil
}

func (c *gzipReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	err := c.zr.Close()
	c.zr = nil
	return err
}

func (c *gzipReader) Reset() {
	c.Close()
}

func shouldDecompressRequest(r *http.Request) bool {
	contentEncoding := r.Header.Values("Content-Encoding")
	return slices.Contains(contentEncoding, "gzip")
}

// GzipMiddleware is an HTTP middleware that handles gzip compression for both requests and responses.
// It compresses responses when the client accepts gzip encoding and decompresses gzip-encoded requests.
func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			gzWriter := gzipWriterPool.Get()
			gzWriter.w = w
			ow = gzWriter
			defer gzipWriterPool.Put(gzWriter)
		}

		if shouldDecompressRequest(r) {
			gzReader := gzipReaderPool.Get()
			err := gzReader.SetUp(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = gzReader
			defer gzipReaderPool.Put(gzReader)
		}

		h.ServeHTTP(ow, r)
	}
}
