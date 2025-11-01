package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestID is a middleware that injects a request ID into the context of each request
// We use Chi's built-in RequestID middleware for compatibility
func RequestID() func(next http.Handler) http.Handler {
	return middleware.RequestID
}
