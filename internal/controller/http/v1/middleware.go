package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mxmntv/anti_bruteforce/pkg/logger"
)

func loggingMiddleware(logger logger.LogInterface, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fmt.Println(start)
		next.ServeHTTP(w, r)
		logline := fmt.Sprintf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
		logger.Info(logline)
	})
}
