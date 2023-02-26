package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mxmntv/anti_bruteforce/pkg/logger"
)

type middleware struct {
	logger logger.LogInterface
}

func newMiddleware(l logger.LogInterface) middleware {
	return middleware{logger: l}
}

func (m middleware) loggingMD(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logline := fmt.Sprintf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
		m.logger.Info(logline)
	})
}

func (m middleware) checkMethodMD(method string, next http.Handler) http.Handler { //nolint:unparam
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			m.logger.Warn(fmt.Sprintf("%s %s", r.Method, r.RequestURI))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}
