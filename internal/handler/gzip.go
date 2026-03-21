package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
)

type gzipWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

func getGzipWriter(w http.ResponseWriter) *gzipWriter {
	writer := gzipWriterPool.Get().(*gzip.Writer)
	writer.Reset(w)
	return &gzipWriter{
		w:  w,
		zw: writer,
	}
}

func (c *gzipWriter) Header() http.Header {
	return c.w.Header()
}

func (c *gzipWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *gzipWriter) Close() error {
	err := c.zw.Close()
	gzipWriterPool.Put(c.zw)
	return err
}

type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

var gzipReaderPool = sync.Pool{
	New: func() interface{} {
		return new(gzip.Reader)
	},
}

func getGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr := gzipReaderPool.Get().(*gzip.Reader)
	if err := zr.Reset(r); err != nil {
		return nil, err
	}
	return &gzipReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *gzipReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	gzipReaderPool.Put(c.zr)
	err := c.zr.Close()
	c.zr = nil
	return err
}

func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			gzWriter := getGzipWriter(w)
			ow = gzWriter
			defer gzWriter.Close()
		}

		contentEncoding := r.Header.Values("Content-Encoding")
		contentType := r.Header.Values("Content-Type")
		sendsGzip := slices.Contains(contentEncoding, "gzip") && slices.Contains(contentType, "application/json")
		if sendsGzip {
			var err error
			gzReader, err := getGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = gzReader
			defer gzReader.Close()
		}

		h.ServeHTTP(ow, r)
	}
}
