package middleware

import (
	"log"
	"net/http"

	"github.com/sergehall/lavoval/apps/api/internal/httpx"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v", rec)
				httpx.Error(w, http.StatusInternalServerError, "internal_error", "Unexpected server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
