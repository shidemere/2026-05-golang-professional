package internalhttp

import (
	"net"
	"net/http"
	"time"
)

type statusResponseWrapper struct {
	http.ResponseWriter
	statusCode int
}

// Wrapper around writing status code for getting it in middleware.
// When it will be called in handler - this will work because of polymorfism.
func (s *statusResponseWrapper) WriteHeader(code int) {
	s.statusCode = code
	s.ResponseWriter.WriteHeader(code)
}

// Golang by default set 200 if something was written, so we need to do it to.
func (s *statusResponseWrapper) Write(b []byte) (int, error) {
	if s.statusCode == 0 {
		s.statusCode = 200
	}
	return s.ResponseWriter.Write(b)
}

func loggingMiddleware(logger Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			logger.Info("can't read ip: ", err)
			host = "unknown"
		}

		wrapper := statusResponseWrapper{
			ResponseWriter: w,
		}

		next(&wrapper, r)

		logger.Info(
			"request handled",
			"ip", host,
			"time", start,
			"method", r.Method,
			"path", r.RequestURI,
			"proto", r.Proto,
			"status", wrapper.statusCode,
			"latency", time.Since(start),
			"user_agent", r.UserAgent(),
		)
	}
}
