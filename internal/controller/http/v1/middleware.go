package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mxmntv/anti_bruteforce/pkg/logger"
)

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{w, http.StatusOK}
}

func loggingMiddleware(logger logger.LogInterface, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)
		lat := time.Since(start)
		logline := fmt.Sprintf("%s [%s] %s %s %s %d %d %s", r.RemoteAddr, start.Format("02/Jan/2006:15:04:05 -0700"),
			r.Method, r.RequestURI, r.Proto, lrw.statusCode, lat, r.UserAgent())
		logger.Info(logline)
	})
}
