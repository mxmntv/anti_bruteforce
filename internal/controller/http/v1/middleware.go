package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mxmntv/anti_bruteforce/pkg/logger"
)

func loggingMD(logger logger.LogInterface, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logline := fmt.Sprintf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
		logger.Info(logline)
	})
}

func checkMethodMD(method string, next http.Handler) http.Handler { //nolint:unparam
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}
